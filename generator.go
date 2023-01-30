package gohateoas

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/survivorbat/go-tsyncmap"
	"reflect"
	"regexp"
	"strings"
)

// typeCacheMap is used to easily fetch json keys from a type
var typeCacheMap = &tsyncmap.Map[string, map[string]string]{}

// iKind is an abstraction of reflect.Value and reflect.Type that allows us to make ensureConcrete generic.
type iKind[T any] interface {
	Kind() reflect.Kind
	Elem() T
}

// ensureConcrete ensures that the given value is a value and not a pointer, if it is, convert it to its element type
func ensureConcrete[T iKind[T]](value T) T {
	if value.Kind() == reflect.Ptr {
		return ensureConcrete[T](value.Elem())
	}

	return value
}

// getFieldNameFromJson returns the field name from the json tag
func getFieldNameFromJson(object any, jsonKey string) (string, error) {
	typeInfo := ensureConcrete(reflect.TypeOf(object))
	if typeInfo.Kind() != reflect.Struct {
		return "", errors.New("object is not a struct")
	}

	typeName := typeNameOf(object)

	// Check for cached values, this way we don't need to perform reflection
	// every time we want to get the field name from a json key.
	if cachedValue, ok := typeCacheMap.Load(typeName); ok {
		return cachedValue[jsonKey], nil
	}

	// It does not
	typeCache := map[string]string{}

	for i := 0; i < typeInfo.NumField(); i++ {
		// Get the json tags of this field
		jsonKey := typeInfo.Field(i).Tag.Get("json")

		// Ignore empty json fields
		if jsonKey == "" || jsonKey == "-" {
			continue
		}

		// Take the first item from the list and use that as the json key, save it
		// in the cache map.
		jsonProperty := strings.Split(jsonKey, ",")[0]
		typeCache[jsonProperty] = typeInfo.Field(i).Name
	}

	typeCacheMap.Store(typeName, typeCache)
	return typeCache[jsonKey], nil
}

// tokenReplaceRegex is a regex that matches tokens in the form of {token}
var tokenReplaceRegex = regexp.MustCompile(`{([^{}]*)}`)

// injectLinks injects the actual links into the struct if any are registered
func injectLinks(registry LinkRegistry, object any, result map[string]any) {
	// Add links if there are any
	links := registry[typeNameOf(object)]

	if len(links) == 0 {
		return
	}

	linkMap := map[string]LinkInfo{}

	// Loop through every link and inject it into the object, replacing tokens
	// with the appropriate values.
	for linkType, linkInfo := range links {
		// Find matches for tokens in the linkInfo like {id} or {name}
		matches := tokenReplaceRegex.FindAllStringSubmatch(linkInfo.Href, -1)

		for _, match := range matches {
			// Check if the value is in the object, like "id" or "name"
			urlValue, ok := result[match[1]]

			if !ok {
				continue
			}

			// Replace the {name} with the actual value
			matchString := fmt.Sprintf("{%s}", match[1])
			linkInfo.Href = strings.Replace(linkInfo.Href, matchString, fmt.Sprintf("%v", urlValue), -1)
		}

		// Save the linkInfo in the object
		linkMap[linkType] = linkInfo
	}

	// Add the _links property
	result["_links"] = linkMap
}

// walkThroughObject goes through the object and injects links into the structs it comes across
func walkThroughObject(registry LinkRegistry, object any, result any) {
	// Prevent nil pointer dereference
	if result == nil {
		return
	}

	// We use this to dissect the object
	reflectValue := ensureConcrete(reflect.ValueOf(object))

	switch result := result.(type) {
	case []any:
		// Loop through the slice's entries and recursively walk through those objects
		for index := range result {
			walkThroughObject(registry, ensureConcrete(reflectValue.Index(index)).Interface(), result[index])
		}

	case map[string]any:
		// Actually inject links, since this is a struct
		injectLinks(registry, object, result)

		// Loop through the map's entries and recursively walk through those objects
		for jsonKey, value := range result {
			switch resultCastValue := value.(type) {
			case map[string]any, []any:
				fieldName, err := getFieldNameFromJson(object, jsonKey)
				if err != nil {
					continue
				}

				fieldValue := reflectValue.FieldByName(fieldName)
				if !fieldValue.IsValid() {
					continue
				}

				walkThroughObject(registry, fieldValue.Interface(), resultCastValue)
			}
		}
	}
}

// InjectLinks is similar to json.Marshal, but it will inject links into the response if the
// registry has any links for the given type. It does this recursively.
func InjectLinks(registry LinkRegistry, object any) []byte {
	rawResponseJson, _ := json.Marshal(object)

	var resultObject any

	switch ensureConcrete(reflect.ValueOf(object)).Kind() {
	case reflect.Slice, reflect.Struct, reflect.Array:
		_ = json.Unmarshal(rawResponseJson, &resultObject)
		walkThroughObject(registry, object, resultObject)

	default:
		// Prevent unnecessary json.Marshal
		return rawResponseJson
	}

	finalResponse, _ := json.Marshal(resultObject)
	return finalResponse
}
