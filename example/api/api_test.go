package api_test

import (
	"github.com/dreamph/flowy"
	"github.com/dreamph/flowy/example/api"
	"github.com/dreamph/flowy/example/fiberx"
	"github.com/gofiber/fiber/v2"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func BenchmarkFlowy(b *testing.B) {
	apiHandler := api.NewNewApiHandler()
	app := fiber.New()

	// Define the route
	app.Get("/", func(c *fiber.Ctx) error {
		return apiHandler.Do(fiberx.WebCtx(c), nil, nil, func(ctx *flowy.Ctx[api.RequestInfo]) (interface{}, error) {
			return "Hi.", nil
		})
	})

	// Create a request to pass to our handler
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Run the benchmark
	for n := 0; n < b.N; n++ {
		resp, err := app.Test(req)
		if err != nil {
			b.Errorf("Request failed: %v", err)
		}
		body, _ := io.ReadAll(resp.Body)
		if err != nil {
			b.Errorf("Reading response body failed: %v", err)
		}
		if body == nil {
			b.Errorf("Unexpected response body: got %v want %v", string(body), "Hi.")
		}
	}
}

func BenchmarkStandard(b *testing.B) {
	app := fiber.New()

	// Define the route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hi.")
	})

	// Create a request to pass to our handler
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	// Run the benchmark
	for n := 0; n < b.N; n++ {
		resp, err := app.Test(req)
		if err != nil {
			b.Errorf("Request failed: %v", err)
		}
		body, _ := io.ReadAll(resp.Body)
		if err != nil {
			b.Errorf("Reading response body failed: %v", err)
		}
		if body == nil {
			b.Errorf("Unexpected response body: got %v want %v", string(body), "Hi.")
		}
	}
}
