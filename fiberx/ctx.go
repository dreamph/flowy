package fiberx

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime/multipart"
)

type WebCtx struct {
	*fiber.Ctx
}

func (c *WebCtx) Method() string {
	return c.Ctx.Method()
}

func (c *WebCtx) UserContext() context.Context {
	return c.Ctx.UserContext()
}

func (c *WebCtx) SendString(statusCode int, text string) error {
	return c.Ctx.Status(statusCode).SendString(text)
}

func (c *WebCtx) SendStream(stream io.Reader, size ...int) error {
	return c.Ctx.SendStream(stream, size...)
}

func (c *WebCtx) JSON(data interface{}) error {
	return c.Ctx.JSON(data)
}

func (c *WebCtx) BodyParser(out interface{}) error {
	return c.Ctx.BodyParser(out)
}

func (c *WebCtx) FormFile(key string) (*multipart.FileHeader, error) {
	return c.Ctx.FormFile(key)
}

func (c *WebCtx) Get(key string, defaultValue ...string) string {
	return c.Ctx.Get(key, defaultValue...)
}

func (c *WebCtx) Status(statusCode int) {
	c.Ctx.Status(statusCode)
}

func With(c *fiber.Ctx) *WebCtx {
	return &WebCtx{Ctx: c}
}
