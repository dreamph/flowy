package api

import (
	"github.com/dreamph/flowy"
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
