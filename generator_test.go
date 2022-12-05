package gohateoas

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type Cupcake struct {
	ID   int    `json:"id"`
	Name string `json:"name"`

	// Infinite loop
	Bakery *Bakery `json:"bakery"`
}

type Bakery struct {
	ID       int        `json:"id"`
	Cupcake  *Cupcake   `json:"cupcake,omitempty"`
	Cupcakes []*Cupcake `json:"cupcakes,omitempty"`
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
		input    *Bakery
		expected map[string]any
	}{
		"nil": {},
		"empty": {
			input: &Bakery{},
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
			input: &Bakery{
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
			input: &Bakery{
				ID: 234,
				Cupcake: &Cupcake{
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
			input: &Bakery{
				ID: 234,
				Cupcakes: []*Cupcake{
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

			RegisterOn(registry, &Cupcake{},
				Index("/api/v1/cupcakes", "test"),
				Self("/api/v1/cupcakes/{id}", "get itself"),
				Custom("other", LinkInfo{Method: http.MethodGet, Href: "/api/v1/cupcakes/{name}", Comment: "get one by name"}),
				Post("/api/v1/cupcakes", "create a new one"))

			RegisterOn(registry, &Bakery{},
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
		input    []*Bakery
		expected []any
	}{
		"nil": {},
		"empty": {
			input:    []*Bakery{},
			expected: []any{},
		},
		"simple": {
			input: []*Bakery{
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
			input: []*Bakery{
				{
					ID: 234,
					Cupcake: &Cupcake{
						ID:   123,
						Name: "abc",
					},
				},
				{
					ID: 777,
					Cupcake: &Cupcake{
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
			input: []*Bakery{
				{
					ID: 234,
					Cupcakes: []*Cupcake{
						{ID: 1, Name: "a"},
						{ID: 3, Name: "c"},
					},
				},
				{
					ID: 879,
					Cupcakes: []*Cupcake{
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

			RegisterOn(registry, &Cupcake{}, Self("/api/v1/cupcakes/{id}", "get itself"))

			RegisterOn(registry, &Bakery{}, Self("/api/v1/bakeries/{id}", "get a bakery by id"))

			// Act
			result := InjectLinks(registry, testData.input)

			// Assert
			var mapResult []any
			_ = json.Unmarshal(result, &mapResult)
			assert.Equal(t, testData.expected, mapResult)
		})
	}
}

type CheeseStore struct {
	ID      int     `json:"id"`
	Cheeses Cheeses `json:"cheeses"`
}

type Cheeses []*Cheese
type Cheese struct {
	ID int `json:"id"`
}

func TestInjectLinks_CreatesExpectedJsonWithRenamedObject(t *testing.T) {
	t.Parallel()
	// Arrange
	input := &CheeseStore{
		ID: 53,
		Cheeses: Cheeses{
			{ID: 54},
			{ID: 21},
		},
	}

	registry := NewLinkRegistry()

	RegisterOn(registry, &CheeseStore{}, Self("/api/v1/stores/{id}", "get itself"))

	RegisterOn(registry, &Cheese{}, Self("/api/v1/cheeses", "get all cheeses"))
	RegisterOn(registry, &Cheeses{}, Self("/api/v1/cheeses", "get all cheeses"))

	// Act
	result := InjectLinks(registry, input)

	// Assert
	expected := map[string]any{
		"id": float64(53),
		"cheeses": []any{
			map[string]any{
				"id": float64(54),
				"_links": map[string]any{
					"self": map[string]any{"comment": "get all cheeses", "href": "/api/v1/cheeses", "method": "GET"},
				},
			},
			map[string]any{
				"id": float64(21),
				"_links": map[string]any{
					"self": map[string]any{"comment": "get all cheeses", "href": "/api/v1/cheeses", "method": "GET"},
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

func TestInjectLinks_ReturnsJsonOnUnknownType(t *testing.T) {
	t.Parallel()
	// Arrange
	registry := NewLinkRegistry()

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
	type DeepType1 struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	type TestType1 struct {
		Deep *DeepType1 `json:"deep"`
	}

	registry := NewLinkRegistry()

	object := &TestType1{
		Deep: &DeepType1{ID: 23, Name: "test"},
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
	type TestType2 struct {
		Deep *DeepType2 `json:"deep"`
	}

	registry := NewLinkRegistry()

	object := []*TestType2{
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

	object := []string{"a", "b", "c"}

	// Act
	result := InjectLinks(registry, object)

	// Assert
	normalJson, _ := json.Marshal(object)
	assert.Equal(t, normalJson, result)
}

func TestGetFieldNameFromJson_ReturnsExpectedName(t *testing.T) {
	t.Parallel()

	type TestType3 struct {
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
			result, err := getFieldNameFromJson(TestType3{}, testData.jsonKey)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, testData.expected, result)
		})
	}
}

func TestGetFieldNameFromJson_ReturnsEmptyStringOnNonStructType(t *testing.T) {
	t.Parallel()
	// Arrange
	type TestType4s []string

	// Act
	result, err := getFieldNameFromJson(TestType4s{}, "any")

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
