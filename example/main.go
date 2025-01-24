package main

import (
	"fmt"
	"github.com/dreamph/flowy"
	"github.com/dreamph/flowy/example/api"
	"github.com/dreamph/flowy/example/fiberx"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	apiHandler := api.NewNewApiHandler()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return apiHandler.Do(fiberx.WebCtx(c), nil, nil, func(ctx *flowy.Ctx[api.RequestInfo]) (interface{}, error) {
			return "Hi.", nil
		})
	})

	app.Get("/custom-status", func(c *fiber.Ctx) error {
		requestOptions := flowy.WithRequestOptions(
			api.SuccessStatus(201),
		)
		return apiHandler.Do(fiberx.WebCtx(c), nil, requestOptions, func(ctx *flowy.Ctx[api.RequestInfo]) (interface{}, error) {
			return "Hi.", nil
		})
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return apiHandler.Do(fiberx.WebCtx(c), nil, nil, func(ctx *flowy.Ctx[api.RequestInfo]) (interface{}, error) {
			return nil, &api.AppError{ErrCode: "0001", ErrMessage: "Error"}
		})
	})

	app.Post("/simple", func(c *fiber.Ctx) error {
		request := &api.SimpleRequest{}
		return apiHandler.Do(fiberx.WebCtx(c), request, nil, func(ctx *flowy.Ctx[api.RequestInfo]) (interface{}, error) {
			fmt.Println(request.Name)
			return request.Name, nil
		})
	})

	app.Post("/upload", func(c *fiber.Ctx) error {
		request := &api.UploadRequest{}
		requestOptions := flowy.WithRequestOptions(
			api.EnableValidate(true),
		)
		return apiHandler.Do(fiberx.WebCtx(c), request, requestOptions, func(ctx *flowy.Ctx[api.RequestInfo]) (interface{}, error) {
			fmt.Println("name:", request.Name)
			fmt.Println("file1:", request.File1.Filename)
			fmt.Println("file2:", request.File2.Filename)
			return "Success", nil
		})
	})

	err := app.Listen(":3000")
	if err != nil {
		log.Fatal(err)
	}
}
