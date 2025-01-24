package fiberx

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"io"
	"mime/multipart"
)

type FiberWebCtx struct {
	*fiber.Ctx
}

func (c *FiberWebCtx) Method() string {
	return c.Ctx.Method()
}

func (c *FiberWebCtx) UserContext() context.Context {
	return c.Ctx.UserContext()
}

func (c *FiberWebCtx) SendString(statusCode int, text string) error {
	return c.Ctx.Status(statusCode).SendString(text)
}

func (c *FiberWebCtx) SendStream(stream io.Reader, size ...int) error {
	return c.Ctx.SendStream(stream, size...)
}

func (c *FiberWebCtx) JSON(data interface{}) error {
	return c.Ctx.JSON(data)
}

func (c *FiberWebCtx) BodyParser(out interface{}) error {
	return c.Ctx.BodyParser(out)
}

func (c *FiberWebCtx) FormFile(key string) (*multipart.FileHeader, error) {
	return c.Ctx.FormFile(key)
}

func (c *FiberWebCtx) Get(key string, defaultValue ...string) string {
	return c.Ctx.Get(key, defaultValue...)
}

func (c *FiberWebCtx) Status(statusCode int) {
	c.Ctx.Status(statusCode)
}

func WebCtx(c *fiber.Ctx) *FiberWebCtx {
	return &FiberWebCtx{Ctx: c}
}
