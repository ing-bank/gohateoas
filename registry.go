package gohateoas

import (
	"fmt"
	"strings"
)

// DefaultLinkRegistry is the global registry for hateoas links
var DefaultLinkRegistry = NewLinkRegistry()

// replacer is used to strip stars and arrays from the %T output
var replacer = strings.NewReplacer("[", "", "]", "", "*", "")

// typeNameOf returns the %T and also strips any stars and arrays from it, used internally for
// HATEOAS.
func typeNameOf(object any) string {
	return replacer.Replace(fmt.Sprintf("%T", object))
}

// NewLinkRegistry instantiates a new LinkRegistry, only used for testing or when overriding
// the DefaultLinkRegistry.
func NewLinkRegistry() LinkRegistry {
	return make(LinkRegistry)
}

// LinkRegistry allows you to register URLs on objects, populating links in responses.
type LinkRegistry map[string]map[string]LinkInfo

// Register registers links to an object using the DefaultLinkRegistry.
func Register(object any, options ...LinkOption) {
	RegisterOn(DefaultLinkRegistry, object, options...)
}

// RegisterOn registers links to an object in the given registry.
func RegisterOn(linkRegistry LinkRegistry, object any, options ...LinkOption) {
	links := make(map[string]LinkInfo)
	for _, option := range options {
		option(links)
	}

	name := typeNameOf(object)
	linkRegistry[name] = links
}
