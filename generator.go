package gohateoas

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

// typeCacheMap is used to easily fetch json keys from a type
var typeCacheMap = &sync.Map{} // map[string]map[string]string{}

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

// ensureNotASlice Ensures that the given value is not a slice, if it is a slice, we use Elem()
// For example: Type []*string will return string. This one is not generic because it doesn't work
// well with reflect.Value.
func ensureNotASlice(value reflect.Type) reflect.Type {
	result := ensureConcrete(value)

	if result.Kind() == reflect.Slice {
		return ensureNotASlice(result.Elem())
	}

	return result
}

// getFieldNameFromJson returns the field name from the json tag
func getFieldNameFromJson(object any, jsonKey string) (string, error) {
	typeInfo := ensureNotASlice(reflect.TypeOf(object))
	if typeInfo.Kind() != reflect.Struct {
		return "", errors.New("object is not a struct")
	}

	typeName := typeNameOf(object)

	// Check for cached values, this way we don't need to perform reflection
	// every time we want to get the field name from a json key.
	if cachedValue, ok := typeCacheMap.Load(typeName); ok {
		return cachedValue.(map[string]string)[jsonKey], nil
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

func injectLinks(registry LinkRegistry, object any, result map[string]any) {
	// Prevent nil pointer dereference
	if result == nil {
		return
	}

	// Add links if there are any
	if links := registry[typeNameOf(object)]; len(links) > 0 {
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

	// Dive deeper into the object and inject links into any nested objects
	valueInfo := ensureConcrete(reflect.ValueOf(object))

	// For deep slices we need to make sure we only take an element of the value, not the entire slice.
	// We also skip everything if a slice is empty.
	if valueInfo.Kind() == reflect.Slice && valueInfo.Len() > 0 {
		valueInfo = ensureConcrete(valueInfo.Index(0))
	} else if valueInfo.Kind() == reflect.Slice {
		return
	}

	for jsonKey, value := range result {
		switch castValue := value.(type) {
		// If the value is a map, we need to inject links into the nested object
		case map[string]any:
			// We retrieve the value of the nested object and recursively call injectLinks
			fieldName, err := getFieldNameFromJson(object, jsonKey)
			if err != nil {
				continue
			}

			fieldValue := valueInfo.FieldByName(fieldName)
			if !fieldValue.IsValid() {
				continue
			}

			injectLinks(registry, fieldValue.Interface(), castValue)

		// If the value is a slice, we need to inject links into each object in the slice
		case []any:
			// We retrieve the value of the nested object and recursively call injectLinks
			fieldName, err := getFieldNameFromJson(object, jsonKey)
			if err != nil {
				continue
			}

			fieldValue := valueInfo.FieldByName(fieldName)
			if !fieldValue.IsValid() {
				continue
			}

			for _, item := range castValue {
				// Only process items that are of type map[string]any
				switch castItem := (item).(type) {
				case map[string]any:
					injectLinks(registry, fieldValue.Interface(), castItem)
				}
			}
		}
	}
}

// InjectLinks is similar to json.Marshal, but it will inject links into the response if the
// registry has any links for the given type. It does this recursively.
func InjectLinks(registry LinkRegistry, object any) []byte {
	rawResponseJson, _ := json.Marshal(object)

	reflectValue := ensureConcrete(reflect.ValueOf(object))

	var resultObject any

	switch reflectValue.Kind() {
	// If the object is a slice, we need to unmarshal it into a slice of maps and use those
	// to inject links into every item
	case reflect.Slice:
		var injectionSlice []any
		_ = json.Unmarshal(rawResponseJson, &injectionSlice)

		// Guard against non-struct slice items, []any can be mixed too, so we have to do it per object
		for index := range injectionSlice {
			switch castItem := (injectionSlice[index]).(type) {
			case map[string]any:
				injectLinks(registry, ensureConcrete(reflectValue.Index(index)).Interface(), castItem)
			}
		}

		resultObject = injectionSlice

	// If the object is a map, we need to unmarshal it and inject the links directly
	case reflect.Struct:
		var injectionMap map[string]any
		_ = json.Unmarshal(rawResponseJson, &injectionMap)

		injectLinks(registry, object, injectionMap)

		resultObject = injectionMap

	// Anything else falls back to a simple unmarshall
	default:
		return rawResponseJson
	}

	finalResponse, _ := json.Marshal(resultObject)
	return finalResponse
}
