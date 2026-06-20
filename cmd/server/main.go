package main

import (
	"log"
	"net/http"
	"os"
	"strings"

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

	cors := core.DefaultCORSConfig()
	if origins := os.Getenv("CORS_ALLOWED_ORIGINS"); origins != "" {
		cors.AllowedOrigins = strings.Split(origins, ",")
	}
	app.EnableCORS(cors)

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", app.Handler()))
}
