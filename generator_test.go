package gohateoas

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type cupcake struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	// Infinite loop
	Bakery *bakery `json:"bakery"`
}

type bakery struct {
	ID       int        `json:"id"`
	Cupcake  *cupcake   `json:"cupcake,omitempty"`
	Cupcakes []*cupcake `json:"cupcakes,omitempty"`
}

func TestTokenReplaceRegex_MatchesCorrectly(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input          string
		expectedGroups [][]string
	}{
		"no groups": {
			input: "/test",
		},
		"id": {
			input:          "/test/{id}",
			expectedGroups: [][]string{{"{id}", "id"}},
		},
		"name and pcode": {
			input:          "/test/{name}/{pcode}",
			expectedGroups: [][]string{{"{name}", "name"}, {"{pcode}", "pcode"}},
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			result := tokenReplaceRegex.FindAllStringSubmatch(testData.input, -1)

			// Assert
			assert.Equal(t, testData.expectedGroups, result)
		})
	}
}

func TestInjectLinks_CreatesExpectedJsonWithObject(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input    *bakery
		expected map[string]any
	}{
		"nil": {},
		"empty": {
			input: &bakery{},
			expected: map[string]any{
				"_links": map[string]any{
					"index": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries",
						"comment": "get all bakeries",
					},
					"self": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries/0",
						"comment": "get a bakery by id",
					},
					"post": map[string]any{
						"method":  "POST",
						"href":    "/api/v1/bakeries",
						"comment": "create a new bakery",
					},
				},
				"id": float64(0),
			},
		},
		"simple": {
			input: &bakery{
				ID: 234,
			},
			expected: map[string]any{
				"id": float64(234),
				"_links": map[string]any{
					"index": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries",
						"comment": "get all bakeries",
					},
					"self": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries/234",
						"comment": "get a bakery by id",
					},
					"post": map[string]any{
						"method":  "POST",
						"href":    "/api/v1/bakeries",
						"comment": "create a new bakery",
					},
				},
			},
		},
		"deep object": {
			input: &bakery{
				ID: 234,
				Cupcake: &cupcake{
					ID:   123,
					Name: "abc",
				},
			},
			expected: map[string]any{
				"id": float64(234),
				"cupcake": map[string]any{
					"name":   "abc",
					"id":     float64(123),
					"bakery": nil,
					"_links": map[string]any{
						"index": map[string]any{"comment": "test", "href": "/api/v1/cupcakes", "method": "GET"},
						"other": map[string]any{"comment": "get one by name", "href": "/api/v1/cupcakes/abc", "method": "GET"},
						"post":  map[string]any{"comment": "create a new one", "href": "/api/v1/cupcakes", "method": "POST"},
						"self":  map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/123", "method": "GET"},
					},
				},
				"_links": map[string]any{
					"index": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries",
						"comment": "get all bakeries",
					},
					"self": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries/234",
						"comment": "get a bakery by id",
					},
					"post": map[string]any{
						"method":  "POST",
						"href":    "/api/v1/bakeries",
						"comment": "create a new bakery",
					},
				},
			},
		},
		"deep array": {
			input: &bakery{
				ID: 234,
				Cupcakes: []*cupcake{
					{ID: 1, Name: "a"},
					{ID: 3, Name: "c"},
				},
			},
			expected: map[string]any{
				"id": float64(234),
				"cupcakes": []any{
					map[string]any{
						"name":   "a",
						"id":     float64(1),
						"bakery": nil,
						"_links": map[string]any{
							"index": map[string]any{"comment": "test", "href": "/api/v1/cupcakes", "method": "GET"},
							"other": map[string]any{"comment": "get one by name", "href": "/api/v1/cupcakes/a", "method": "GET"},
							"post":  map[string]any{"comment": "create a new one", "href": "/api/v1/cupcakes", "method": "POST"},
							"self":  map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/1", "method": "GET"},
						},
					},
					map[string]any{
						"name":   "c",
						"id":     float64(3),
						"bakery": nil,
						"_links": map[string]any{
							"index": map[string]any{"comment": "test", "href": "/api/v1/cupcakes", "method": "GET"},
							"other": map[string]any{"comment": "get one by name", "href": "/api/v1/cupcakes/c", "method": "GET"},
							"post":  map[string]any{"comment": "create a new one", "href": "/api/v1/cupcakes", "method": "POST"},
							"self":  map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/3", "method": "GET"},
						},
					},
				},
				"_links": map[string]any{
					"index": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries",
						"comment": "get all bakeries",
					},
					"self": map[string]any{
						"method":  "GET",
						"href":    "/api/v1/bakeries/234",
						"comment": "get a bakery by id",
					},
					"post": map[string]any{
						"method":  "POST",
						"href":    "/api/v1/bakeries",
						"comment": "create a new bakery",
					},
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

			RegisterOn(registry, &cupcake{},
				Index("/api/v1/cupcakes", "test"),
				Self("/api/v1/cupcakes/{id}", "get itself"),
				Custom("other", LinkInfo{Method: http.MethodGet, Href: "/api/v1/cupcakes/{name}", Comment: "get one by name"}),
				Post("/api/v1/cupcakes", "create a new one"))

			RegisterOn(registry, &bakery{},
				Index("/api/v1/bakeries", "get all bakeries"),
				Self("/api/v1/bakeries/{id}", "get a bakery by id"),
				Post("/api/v1/bakeries", "create a new bakery"))

			// Act
			result := InjectLinks(registry, testData.input)

			// Assert
			var mapResult map[string]any
			_ = json.Unmarshal(result, &mapResult)
			assert.Equal(t, testData.expected, mapResult)
		})
	}
}

func TestInjectLinks_CreatesExpectedJsonWithSlice(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input    []*bakery
		expected []any
	}{
		"nil": {},
		"empty": {
			input:    []*bakery{},
			expected: []any{},
		},
		"simple": {
			input: []*bakery{
				{ID: 234},
				{ID: 556},
			},
			expected: []any{
				map[string]any{
					"id": float64(234),
					"_links": map[string]any{
						"self": map[string]any{
							"method":  "GET",
							"href":    "/api/v1/bakeries/234",
							"comment": "get a bakery by id",
						},
					},
				},
				map[string]any{
					"id": float64(556),
					"_links": map[string]any{
						"self": map[string]any{
							"method":  "GET",
							"href":    "/api/v1/bakeries/556",
							"comment": "get a bakery by id",
						},
					},
				},
			},
		},
		"deep object": {
			input: []*bakery{
				{
					ID: 234,
					Cupcake: &cupcake{
						ID:   123,
						Name: "abc",
					},
				},
				{
					ID: 777,
					Cupcake: &cupcake{
						ID:   88,
						Name: "abc",
					},
				},
			},
			expected: []any{
				map[string]any{
					"id": float64(234),
					"cupcake": map[string]any{
						"name":   "abc",
						"id":     float64(123),
						"bakery": nil,
						"_links": map[string]any{
							"self": map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/123", "method": "GET"},
						},
					},
					"_links": map[string]any{
						"self": map[string]any{
							"method":  "GET",
							"href":    "/api/v1/bakeries/234",
							"comment": "get a bakery by id",
						},
					},
				},
				map[string]any{
					"id": float64(777),
					"cupcake": map[string]any{
						"name":   "abc",
						"id":     float64(88),
						"bakery": nil,
						"_links": map[string]any{
							"self": map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/88", "method": "GET"},
						},
					},
					"_links": map[string]any{
						"self": map[string]any{
							"method":  "GET",
							"href":    "/api/v1/bakeries/777",
							"comment": "get a bakery by id",
						},
					},
				},
			},
		},
		"deep array": {
			input: []*bakery{
				{
					ID: 234,
					Cupcakes: []*cupcake{
						{ID: 1, Name: "a"},
						{ID: 3, Name: "c"},
					},
				},
				{
					ID: 879,
					Cupcakes: []*cupcake{
						{ID: 5, Name: "a"},
						{ID: 6, Name: "c"},
					},
				},
			},
			expected: []any{
				map[string]any{
					"id": float64(234),
					"cupcakes": []any{
						map[string]any{
							"name":   "a",
							"id":     float64(1),
							"bakery": nil,
							"_links": map[string]any{
								"self": map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/1", "method": "GET"},
							},
						},
						map[string]any{
							"name":   "c",
							"id":     float64(3),
							"bakery": nil,
							"_links": map[string]any{
								"self": map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/3", "method": "GET"},
							},
						},
					},
					"_links": map[string]any{
						"self": map[string]any{
							"method":  "GET",
							"href":    "/api/v1/bakeries/234",
							"comment": "get a bakery by id",
						},
					},
				},
				map[string]any{
					"id": float64(879),
					"cupcakes": []any{
						map[string]any{
							"name":   "a",
							"id":     float64(5),
							"bakery": nil,
							"_links": map[string]any{
								"self": map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/5", "method": "GET"},
							},
						},
						map[string]any{
							"name":   "c",
							"id":     float64(6),
							"bakery": nil,
							"_links": map[string]any{
								"self": map[string]any{"comment": "get itself", "href": "/api/v1/cupcakes/6", "method": "GET"},
							},
						},
					},
					"_links": map[string]any{
						"self": map[string]any{
							"method":  "GET",
							"href":    "/api/v1/bakeries/879",
							"comment": "get a bakery by id",
						},
					},
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

			RegisterOn(registry, &cupcake{}, Self("/api/v1/cupcakes/{id}", "get itself"))
			RegisterOn(registry, &bakery{}, Self("/api/v1/bakeries/{id}", "get a bakery by id"))

			// Act
			result := InjectLinks(registry, testData.input)

			// Assert
			var mapResult []any
			_ = json.Unmarshal(result, &mapResult)
			assert.Equal(t, testData.expected, mapResult)
		})
	}
}

// Deep slices originally didn't work, so this test is to ensure that they do.

type cheeseStore struct {
	ID      int           `json:"id"`
	Cheeses [][][]*cheese `json:"cheeses"`
}

type cheese struct {
	ID int `json:"id"`
}

func TestInjectLinks_CreatesExpectedJsonWithDeeperSlice(t *testing.T) {
	t.Parallel()
	// Arrange
	input := &cheeseStore{
		ID: 53,
		Cheeses: [][][]*cheese{
			{
				{
					{
						ID: 54,
					},
					{
						ID: 21,
					},
				},
			},
		},
	}

	registry := NewLinkRegistry()

	RegisterOn(registry, &cheeseStore{}, Self("/api/v1/stores/{id}", "get itself"))
	RegisterOn(registry, &cheese{}, Index("/api/v1/cheeses", "get all cheeses"))

	// Act
	result := InjectLinks(registry, input)

	// Assert
	expected := map[string]any{
		"id": float64(53),
		"cheeses": []any{
			[]any{
				[]any{
					map[string]any{
						"id": float64(54),
						"_links": map[string]any{
							"index": map[string]any{"comment": "get all cheeses", "href": "/api/v1/cheeses", "method": "GET"},
						},
					},
					map[string]any{
						"id": float64(21),
						"_links": map[string]any{
							"index": map[string]any{"comment": "get all cheeses", "href": "/api/v1/cheeses", "method": "GET"},
						},
					},
				},
			},
		},
		"_links": map[string]any{
			"self": map[string]any{
				"method":  "GET",
				"href":    "/api/v1/stores/53",
				"comment": "get itself",
			},
		},
	}

	var mapResult map[string]any
	_ = json.Unmarshal(result, &mapResult)
	assert.Equal(t, expected, mapResult)
}

// empty is used as a dummy to make sure the if-statement in InjectLinks doesn't halt execution on no registered links
type empty struct{}

func TestInjectLinks_ReturnsJsonOnUnknownType(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewLinkRegistry()

	RegisterOn(registry, empty{}, Self("", ""))

	// Act
	result := InjectLinks(registry, "test")

	// Assert
	normalJson, _ := json.Marshal("test")
	assert.Equal(t, normalJson, result)
}

// Test objects are numbered because of the caching mechanism.

func TestInjectLinks_IgnoresIfNoTypeRegistered(t *testing.T) {
	t.Parallel()
	// Arrange
	type deepType1 struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type testType1 struct {
		Deep *deepType1 `json:"deep"`
	}

	registry := NewLinkRegistry()
	RegisterOn(registry, empty{}, Self("", ""))

	object := &testType1{
		Deep: &deepType1{ID: 23, Name: "test"},
	}

	// Act
	result := InjectLinks(registry, object)

	// Assert
	normalJson, _ := json.Marshal(object)
	assert.Equal(t, normalJson, result)
}

func TestInjectLinks_IgnoresIfNoTypeRegisteredOnSlice(t *testing.T) {
	t.Parallel()
	// Arrange
	type DeepType2 struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type testType2 struct {
		Deep *DeepType2 `json:"deep"`
	}

	registry := NewLinkRegistry()
	RegisterOn(registry, empty{}, Self("", ""))

	object := []*testType2{
		{Deep: &DeepType2{ID: 23, Name: "test"}},
		{Deep: &DeepType2{ID: 7, Name: "other"}},
	}

	// Act
	result := InjectLinks(registry, object)

	// Assert
	normalJson, _ := json.Marshal(object)
	assert.Equal(t, normalJson, result)
}

func TestInjectLinks_IgnoresOnNonStructSlices(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewLinkRegistry()
	RegisterOn(registry, empty{}, Self("", ""))

	object := []string{"a", "b", "c"}

	// Act
	result := InjectLinks(registry, object)

	// Assert
	normalJson, _ := json.Marshal(object)
	assert.Equal(t, normalJson, result)
}

func TestInjectLinks_IgnoresIfRegistryIsEmpty(t *testing.T) {
	t.Parallel()
	// Arrange
	type testType3 struct{}

	registry := NewLinkRegistry()

	object := &testType3{}

	// Act
	result := InjectLinks(registry, object)

	// Assert
	normalJson, _ := json.Marshal(object)
	assert.Equal(t, normalJson, result)
}

func TestGetFieldNameFromJson_ReturnsExpectedName(t *testing.T) {
	t.Parallel()

	type testType3 struct {
		Name string `json:"name"`
		Deep int    `json:"deep,omitempty"`
	}

	tests := map[string]struct {
		jsonKey  string
		expected string
	}{
		"name": {
			jsonKey:  "name",
			expected: "Name",
		},
		"deep": {
			jsonKey:  "deep",
			expected: "Deep",
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			result, err := getFieldNameFromJson(testType3{}, testData.jsonKey)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, testData.expected, result)
		})
	}
}

func TestGetFieldNameFromJson_ReturnsEmptyStringOnNonStructType(t *testing.T) {
	t.Parallel()
	// Arrange
	type testType4s []string

	// Act
	result, err := getFieldNameFromJson(testType4s{}, "any")

	// Assert
	assert.EqualError(t, err, "object is not a struct")
	assert.Equal(t, "", result)
}

func TestGetFieldNameFromJson_IgnoresMissingJsonFields(t *testing.T) {
	t.Parallel()

	type OtherType1 struct {
		Name string `json:""`
		Deep int    `json:"-"`
	}

	tests := map[string]struct {
		jsonKey  string
		expected string
	}{
		"name": {
			jsonKey: "name",
		},
		"deep": {
			jsonKey: "deep",
		},
	}

	for name, testData := range tests {
		testData := testData
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			// Act
			result, err := getFieldNameFromJson(OtherType1{}, testData.jsonKey)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, "", result)
		})
	}
}

func TestEnsureConcrete_TurnsTypeTestAIntoValue(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestA struct{}
	reflectPointer := reflect.TypeOf(&TestA{})

	// Act
	result := ensureConcrete(reflectPointer)

	// Assert
	reflectValue := reflect.TypeOf(TestA{})

	assert.Equal(t, reflectValue, result)
}

func TestEnsureConcrete_TurnsTypeTestBIntoValue(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestB struct{}
	reflectPointer := reflect.TypeOf(&TestB{})

	// Act
	result := ensureConcrete(reflectPointer)

	// Assert
	reflectValue := reflect.TypeOf(TestB{})

	assert.Equal(t, reflectValue, result)
}

func TestEnsureConcrete_TurnsTypeTestAIntoValueWithMultiplePointers(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestA struct{}

	first := &TestA{}
	second := &first
	third := &second

	reflectPointer := reflect.TypeOf(&third)

	// Act
	result := ensureConcrete(reflectPointer)

	// Assert
	reflectValue := reflect.TypeOf(TestA{})

	assert.Equal(t, reflectValue, result)
}

func TestEnsureConcrete_LeavesValueOfTypeTestAAlone(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestA struct{}
	reflectValue := reflect.TypeOf(TestA{})

	// Act
	result := ensureConcrete(reflectValue)

	// Assert
	assert.Equal(t, reflectValue, result)
}

func TestEnsureConcrete_LeavesValueOfTypeTestBAlone(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestB struct{}
	reflectValue := reflect.TypeOf(TestB{})

	// Act
	result := ensureConcrete(reflectValue)

	// Assert
	assert.Equal(t, reflectValue, result)
}

func BenchmarkInjectLinks(b *testing.B) {
	type base struct {
		ID             [16]byte  `json:"id"`
		Name           string    `json:"name"`
		Brand          string    `json:"brand"`
		Weight         int       `json:"weight"`
		Expired        bool      `json:"expired"`
		ExpirationDate time.Time `json:"expirationDate"`
	}

	type fruit struct {
		base
	}

	type vegetable struct {
		base
	}

	type cake struct {
		base
	}

	type fridge struct {
		base

		Cakes      []cake
		Fruits     []fruit
		Vegetables []vegetable
	}

	inputTests := map[string]func() []fridge{
		"1 fridge with 1 of each": func() []fridge {
			return []fridge{
				{
					Cakes:      make([]cake, 1),
					Fruits:     make([]fruit, 1),
					Vegetables: make([]vegetable, 1),
				},
			}
		},
		"1 fridge with 6000 of each": func() []fridge {
			return []fridge{
				{
					Cakes:      make([]cake, 6000),
					Fruits:     make([]fruit, 6000),
					Vegetables: make([]vegetable, 6000),
				},
			}
		},
		"6000 empty fridges": func() []fridge {
			return make([]fridge, 6000)
		},
		"6000 fridges with 1 of each": func() []fridge {
			result := make([]fridge, 6000)

			for i := 0; i < len(result); i++ {
				result[i] = fridge{
					Cakes:      make([]cake, 1),
					Fruits:     make([]fruit, 1),
					Vegetables: make([]vegetable, 1),
				}
			}

			return result
		},
		"600 fridges with 600 of each": func() []fridge {
			result := make([]fridge, 600)

			for i := 0; i < len(result); i++ {
				result[i] = fridge{
					Cakes:      make([]cake, 600),
					Fruits:     make([]fruit, 600),
					Vegetables: make([]vegetable, 600),
				}
			}

			return result
		},
	}

	registryTests := map[string]func() LinkRegistry{
		// This test won't do much because there's an if-statement blocking execution, but it gives us a bit of insight
		"no links": NewLinkRegistry,

		"3 links for fridge": func() LinkRegistry {
			registry := NewLinkRegistry()
			RegisterOn(registry, fridge{}, Self("/api/fridges", "Get this fridge"), Post("/api/fridges", "Create a new fridge"), Delete("/api/v1/fridges/{id}", "Delete a fridge"))

			return registry
		},

		"3 links for all objects": func() LinkRegistry {
			registry := NewLinkRegistry()
			RegisterOn(registry, fridge{}, Self("/api/fridges", "Get this fridge"), Post("/api/fridges", "Create a new fridge"), Delete("/api/v1/fridges/{id}", "Delete a fridge"))
			RegisterOn(registry, vegetable{}, Self("/api/vegetables", "Get this vegetable"), Post("/api/vegetables", "Create a new vegetable"), Delete("/api/v1/vegetables/{id}", "Delete a vegetable"))
			RegisterOn(registry, fruit{}, Self("/api/fruits", "Get this fruit"), Post("/api/fruits", "Create a new fruit"), Delete("/api/v1/fruits/{id}", "Delete a fruit"))
			RegisterOn(registry, cake{}, Self("/api/cakes", "Get this cake"), Post("/api/cakes", "Create a new cake"), Delete("/api/v1/cakes/{id}", "Delete a cake"))

			return registry
		},
	}

	for registryName, registryData := range registryTests {
		for inputName, inputData := range inputTests {
			b.Run(fmt.Sprintf("%s, %s", registryName, inputName), func(b *testing.B) {
				for i := 0; i < b.N; i++ {
					_ = InjectLinks(registryData(), inputData())
				}
			})
		}
	}
}
