package gohateoas

func ExampleRegister() {
	type Cupcake struct{}

	Register(Cupcake{}, Self("/api/v1/cupcakes/{id}", "get a cupcake"),
		Post("/api/v1/cupcakes", "create a cupcake"),
		Put("/api/v1/cupcakes/{id}", "fully update a cupcake"),
		Patch("/api/v1/cupcakes/{id}", "partially update a cupcake"),
		Delete("/api/v1/cupcakes?id={id}", "delete this cupcake"))
}
