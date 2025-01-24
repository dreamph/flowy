## _Flowy (Smooth API flow)_

## Benchmark

```shell
cpu: Apple M4 Pro
BenchmarkFlowy-12       	  300319	      3989 ns/op
BenchmarkStandard-12    	  308863	      3747 ns/op
```

## Basic Usage
Full Example [example](example)

```go
package main

import (
	"fmt"
	"github.com/dreamph/flowy"
	"github.com/dreamph/flowy/fiberx"
	"github.com/gofiber/fiber/v2"
	errs "github.com/pkg/errors"
	"io"
	"log"
	"mime/multipart"
	"time"
)

type RequestInfo struct {
	Token string `json:"token"`
}

type RequestOption struct {
	EnableValidate bool
	SuccessStatus  int
}

func EnableValidate(enable bool) flowy.RequestOptions[RequestOption] {
	return func(opts *RequestOption) {
		opts.EnableValidate = enable
	}
}

func SuccessStatus(successStatus int) flowy.RequestOptions[RequestOption] {
	return func(opts *RequestOption) {
		opts.SuccessStatus = successStatus
	}
}

type ErrorResponse struct {
	Status        bool            `json:"status"`
	StatusCode    int             `json:"statusCode"`
	StatusMessage string          `json:"statusMessage"`
	Type          string          `json:"type"`
	Code          string          `json:"code"`
	Message       string          `json:"message"`
	ErrorMessage  string          `json:"errorMessage"`
	Time          time.Time       `json:"time" swaggertype:"string" format:"date-time"`
	Detail        string          `json:"detail"`
	ErrorData     *[]AppErrorData `json:"errorData"`
	Cause         error           `json:"-"`
}

type AppErrorData struct {
	Reference    string           `json:"reference"`
	ErrorDetails []AppErrorDetail `json:"errorDetails"`
}

type AppErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type StreamData struct {
	Data io.Reader
	Size int `json:"status"`
}

type AppError struct {
	ErrCode    string `json:"errCode"`
	ErrMessage string `json:"errMessage"`
}

func (e *AppError) Error() string {
	return e.ErrCode + ":" + e.ErrMessage
}

func NewApiResponseHandler() flowy.ApiResponseHandler[flowy.WebCtx, RequestOption] {
	apiResponseHandler := flowy.NewApiResponseHandler[flowy.WebCtx, RequestOption](&flowy.ApiResponseHandlerOptions[flowy.WebCtx, RequestOption]{
		ResponseSuccess: func(c flowy.WebCtx, requestOption *RequestOption, data any) error {
			if requestOption.SuccessStatus > 0 {
				c.Status(requestOption.SuccessStatus)
			}

			streamData, ok := data.(*StreamData)
			if ok {
				if streamData.Size > 0 {
					return c.SendStream(streamData.Data, streamData.Size)
				} else {
					return c.SendStream(streamData.Data)
				}
			}
			return c.JSON(data)
		},
		ResponseError: func(c flowy.WebCtx, requestOption *RequestOption, err error) error {
			res := &ErrorResponse{
				Status:     false,
				StatusCode: 500,
				Code:       "E00001",
				Message:    err.Error(),
			}

			var appError *AppError
			ok := errs.As(err, &appError)
			if ok {
				res.Code = appError.ErrCode
				res.Message = appError.ErrMessage
				res.StatusCode = 400
			}
			c.Status(res.StatusCode)
			return c.JSON(res)
		},
	})
	return apiResponseHandler
}

func NewNewApiHandler() flowy.ApiHandler[flowy.WebCtx, RequestInfo, RequestOption] {
	requestValidator := flowy.NewRequestValidator()
	responseHandler := NewApiResponseHandler()
	return flowy.NewApiHandler[flowy.WebCtx, RequestInfo, RequestOption](responseHandler, &flowy.ApiHandlerOptions[flowy.WebCtx, RequestInfo, RequestOption]{
		OnValidate: func(c flowy.WebCtx, requestOption *RequestOption, data any) error {
			if requestOption.EnableValidate {
				err := requestValidator.Validate(data)
				if err != nil {
					return &AppError{ErrCode: "V0001", ErrMessage: err.Error()}
				}
				return nil
			}
			return nil
		},
		OnBefore: func(c flowy.WebCtx, requestOption *RequestOption) error {
			log.Println("OnBefore")
			if requestOption.EnableValidate {
				log.Println("EnableValidate")
			}
			return nil
		},
		GetRequestInfo: func(c flowy.WebCtx, requestOption *RequestOption) (*RequestInfo, error) {
			log.Println("GetRequestInfo")
			return &RequestInfo{
				Token: "my-token",
			}, nil
		},
		OnAfter: func(c flowy.WebCtx, requestOption *RequestOption) error {
			log.Println("OnAfter")
			return nil
		},
	})
}

type UploadRequest struct {
	Name  string                `form:"name"`
	File1 *multipart.FileHeader `form:"file1" validate:"allow-file-extensions=.go1,allow-file-mime-types=text/plain:text/plain2"`
	File2 *multipart.FileHeader `form:"file2"`
}

type SimpleRequest struct {
	Name string `json:"name"`
}

func main() {
	apiHandler := NewNewApiHandler()
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return apiHandler.Do(fiberx.With(c), nil, nil, func(ctx *flowy.Ctx[RequestInfo]) (interface{}, error) {
			return "Hi.", nil
		})
	})

	app.Get("/custom-status", func(c *fiber.Ctx) error {
		requestOptions := flowy.WithRequestOptions(
			SuccessStatus(201),
		)
		return apiHandler.Do(fiberx.With(c), nil, requestOptions, func(ctx *flowy.Ctx[RequestInfo]) (interface{}, error) {
			return "Hi.", nil
		})
	})

	app.Get("/error", func(c *fiber.Ctx) error {
		return apiHandler.Do(fiberx.With(c), nil, nil, func(ctx *flowy.Ctx[RequestInfo]) (interface{}, error) {
			return nil, &AppError{ErrCode: "0001", ErrMessage: "Error"}
		})
	})

	app.Post("/simple", func(c *fiber.Ctx) error {
		request := &SimpleRequest{}
		return apiHandler.Do(fiberx.With(c), request, nil, func(ctx *flowy.Ctx[RequestInfo]) (interface{}, error) {
			fmt.Println(request.Name)
			return request.Name, nil
		})
	})

	app.Post("/upload", func(c *fiber.Ctx) error {
		request := &UploadRequest{}
		requestOptions := flowy.WithRequestOptions(
			EnableValidate(true),
		)
		return apiHandler.Do(fiberx.With(c), request, requestOptions, func(ctx *flowy.Ctx[RequestInfo]) (interface{}, error) {
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


```

## Test

```shell
curl -v http://localhost:3000/
```

```shell
curl -v http://localhost:3000/custom-status
```

```shell
curl -v http://localhost:3000/error
```

```shell
curl -v -X POST -d '{"name": "Hello"}' http://localhost:3000/simple -H 'Content-Type: application/json'
```

```shell
curl -v -F name=cenery -F file1=@api.go -F file2=@utils.go http://localhost:3000/upload
```

Buy Me a Coffee
=======
[!["Buy Me A Coffee"](https://www.buymeacoffee.com/assets/img/custom_images/orange_img.png)](https://www.buymeacoffee.com/dreamph)
