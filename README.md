# 🦁 Golang Hateoas

[![Go package](https://github.com/ing-bank/gohateoas/actions/workflows/test.yaml/badge.svg)](https://github.com/ing-bank/gohateoas/actions/workflows/test.yaml)
![GitHub](https://img.shields.io/github/license/ing-bank/gohateoas)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/ing-bank/gohateoas)

Hateoas isn't always necessary for a REST API to function, but it's definitely a nice-to-have. 
This library provides a simple way to add hateoas links to your API responses, regardless
of the wrapper object you use.

## ⬇️ Installation

`go get github.com/ing-bank/gohateoas`

## 📋 Usage

```go
package main

import (
	"fmt"
	"github.com/ing-bank/gohateoas"
	"encoding/json"
)

// APIResponse is a simple struct that will be used as a wrapper for our response.
type APIResponse struct {
	Data any `json:"data"`
}

// MarshalJSON overrides the usual json.Marshal behaviour to allow us to add links to the response
func (a APIResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Data json.RawMessage `json:"data"`
	}{
		Data: gohateoas.InjectLinks(gohateoas.DefaultLinkRegistry, a.Data),
	})
}

type Cupcake struct{}

func main() {
	gohateoas.Register(Cupcake{}, gohateoas.Self("/api/v1/cupcakes/{id}", "Get this cupcake"),
		gohateoas.Post("/api/v1/cupcakes", "Create a cupcake"),
		gohateoas.Patch("/api/v1/cupcakes/{id}", "Partially update a cupcake"),
		gohateoas.Delete("/api/v1/cupcakes?id={id}", "Delete this cupcake"))

	response := APIResponse{Data: Cupcake{}}

	jsonBytes, _ := json.Marshal(response)

	fmt.Println(string(jsonBytes))
}

```

## 🚀 Development

1. Clone the repository
2. Run `make t` to run unit tests
3. Run `make fmt` to format code
4. Run `make docs` to update the documentation

You can run `make` to see a list of useful commands.

## 🔭 Future Plans

- [ ] Add support for custom link-property name
