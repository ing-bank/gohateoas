package gohateoas

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetTypeName_ReturnsNameOnTypeA(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestTypeA struct{}

	// Act
	result := typeNameOf(new(TestTypeA))

	// Assert
	assert.Equal(t, "gohateoas.TestTypeA", result)
}

func TestGetTypeName_ReturnsNameOnTypeB(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestTypeB struct{}

	// Act
	result := typeNameOf(new(*****[]*[]*[]****TestTypeB))

	// Assert
	assert.Equal(t, "gohateoas.TestTypeB", result)
}

type TestType struct{}

func TestRegisterOn_RegistersExpectedLinks(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		options  []LinkOption
		expected LinkRegistry
	}{
		"no options": {
			options: []LinkOption{},
			expected: map[string]map[string]LinkInfo{
				"gohateoas.TestType1": {},
			},
		},
		"all options": {
			options: []LinkOption{
				Self("/cupcakes/{id}", "Get a single cupcake"),
				Index("/cupcakes", "Get all cupcakes"),
				Post("/cupcakes", "Create a new cupcake"),
				Put("/cupcakes/{id}", "Fully update a cupcake"),
				Patch("/cupcakes/{id}", "Partially update a cupcake"),
				Delete("/cupcakes", "Delete a cupcake"),
				Custom("custom", LinkInfo{Method: http.MethodConnect, Href: "/cupcakes/custom", Comment: "Custom action"}),
			},
			expected: map[string]map[string]LinkInfo{
				"gohateoas.TestType1": {
					"self":   LinkInfo{Method: http.MethodGet, Href: "/cupcakes/{id}", Comment: "Get a single cupcake"},
					"index":  LinkInfo{Method: http.MethodGet, Href: "/cupcakes", Comment: "Get all cupcakes"},
					"post":   LinkInfo{Method: http.MethodPost, Href: "/cupcakes", Comment: "Create a new cupcake"},
					"put":    LinkInfo{Method: http.MethodPut, Href: "/cupcakes/{id}", Comment: "Fully update a cupcake"},
					"patch":  LinkInfo{Method: http.MethodPatch, Href: "/cupcakes/{id}", Comment: "Partially update a cupcake"},
					"delete": LinkInfo{Method: http.MethodDelete, Href: "/cupcakes", Comment: "Delete a cupcake"},
					"custom": LinkInfo{Method: http.MethodConnect, Href: "/cupcakes/custom", Comment: "Custom action"},
				},
			},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Arrange
			registry := NewLinkRegistry()

			// Act
			RegisterOn(registry, TestType{}, testData.options...)

			// Assert
			assert.Equal(t, testData.expected, registry)
		})
	}
}

func TestRegister_UsesDefaultRegistry(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestType struct{}

	// Act
	Register(TestType{}, Self("test", "get it"))

	// Assert
	assert.NotEmpty(t, DefaultLinkRegistry["gohateoas.TestType1"])
}
