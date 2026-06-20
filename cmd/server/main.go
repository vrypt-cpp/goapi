package main

import (
	"log"
	"net/http"

	"goapi/core"
	"goapi/plugins/health"
	"goapi/plugins/users"
)

func main() {
	app := core.NewApp("GoAPI", "1.0.0", "/api/v1")

	app.Use(
		health.New(),
		users.New(),
	)

	app.ServeOpenAPI("/openapi.json")

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.Handler()))
}
