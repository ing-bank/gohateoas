package gohateoas

import (
	"encoding/json"
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

// getFieldNameFromJson returns the field name from the json tag
func getFieldNameFromJson(object any, jsonKey string) string {
	typeInfo := ensureConcrete(reflect.TypeOf(object))
	if typeInfo.Kind() != reflect.Struct {
		return ""
	}

	typeName := typeNameOf(object)

	// Check for cached values, this way we don't need to perform reflection
	// every time we want to get the field name from a json key.
	if cachedValue, ok := typeCacheMap.Load(typeName); ok {
		return cachedValue.(map[string]string)[jsonKey]
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
	return typeCache[jsonKey]
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
	typeInfo := ensureConcrete(reflect.ValueOf(object))

	for jsonKey, value := range result {
		switch castValue := (value).(type) {
		// If the value is a map, we need to inject links into the nested object
		case map[string]any:
			// We retrieve the value of the nested object and recursively call injectLinks
			fieldValue := typeInfo.FieldByName(getFieldNameFromJson(object, jsonKey))
			if !fieldValue.IsValid() {
				continue
			}

			injectLinks(registry, fieldValue.Interface(), castValue)

		// If the value is a slice, we need to inject links into each object in the slice
		case []any:
			// We retrieve the value of the nested object and recursively call injectLinks
			fieldValue := typeInfo.FieldByName(getFieldNameFromJson(object, jsonKey))
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
				injectLinks(registry, reflectValue.Index(index).Interface(), castItem)
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
