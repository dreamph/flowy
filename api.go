package flowy

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
)

type WebCtx interface {
	Method() string
	UserContext() context.Context
	SendString(statusCode int, text string) error
	SendStream(stream io.Reader, size ...int) error
	JSON(data interface{}) error
	BodyParser(out interface{}) error
	FormFile(key string) (*multipart.FileHeader, error)
	Get(key string, defaultValue ...string) string
	Status(status int)
}

type RequestOptions[T any] func(opts *T)

func WithRequestOptions[T any](opts ...RequestOptions[T]) *T {
	var opt T
	if opts == nil {
		return &opt
	}
	for _, o := range opts {
		o(&opt)
	}
	return &opt
}

type DoFunc[RequestInfo any] func(ctx *Ctx[RequestInfo]) (interface{}, error)

type apiHandler[WCtx WebCtx, RequestInfo any, RequestOption any] struct {
	apiResponseHandler ApiResponseHandler[WCtx, RequestOption]
	options            *ApiHandlerOptions[WCtx, RequestInfo, RequestOption]
}

type ApiHandlerOptions[WCtx WebCtx, RequestInfo any, RequestOption any] struct {
	GetRequestInfo func(c WCtx, requestOption *RequestOption) (*RequestInfo, error)
	OnBefore       func(c WCtx, requestOption *RequestOption) error
	OnValidate     func(c WCtx, requestOption *RequestOption, data any) error
	OnAfter        func(c WCtx, requestOption *RequestOption) error
}

func NewApiHandler[WCtx WebCtx, RequestInfo any, RequestOption any](apiResponseHandler ApiResponseHandler[WCtx, RequestOption], options *ApiHandlerOptions[WCtx, RequestInfo, RequestOption]) ApiHandler[WCtx, RequestInfo, RequestOption] {
	return &apiHandler[WCtx, RequestInfo, RequestOption]{
		apiResponseHandler: apiResponseHandler,
		options:            options,
	}
}

type ApiHandler[WCtx WebCtx, RequestInfo any, RequestOption any] interface {
	Do(c WCtx, requestPtr interface{}, requestOption *RequestOption, doFunc DoFunc[RequestInfo]) error
}

func (h *apiHandler[WCtx, RequestInfo, RequestOption]) defaultRequestOptionIfNull(requestOption *RequestOption) *RequestOption {
	if requestOption != nil {
		return requestOption
	}

	var opt RequestOption
	return &opt
}

func (h *apiHandler[WCtx, RequestInfo, RequestOption]) Do(c WCtx, requestPtr any, requestOption *RequestOption, doFunc DoFunc[RequestInfo]) error {
	requestOption = h.defaultRequestOptionIfNull(requestOption)

	err := h.options.OnBefore(c, requestOption)
	if err != nil {
		return h.apiResponseHandler.ResponseError(c, requestOption, err)
	}

	_, err = h.bodyParserIfRequired(c, requestOption, requestPtr)
	if err != nil {
		return h.apiResponseHandler.ResponseError(c, requestOption, err)
	}

	if h.options.OnValidate != nil {
		err = h.options.OnValidate(c, requestOption, requestPtr)
		if err != nil {
			return h.apiResponseHandler.ResponseError(c, requestOption, err)
		}
	}

	requestInfo, err := h.options.GetRequestInfo(c, requestOption)
	if err != nil {
		return h.apiResponseHandler.ResponseError(c, requestOption, err)
	}

	data, err := doFunc(&Ctx[RequestInfo]{Context: c.UserContext(), RequestInfo: requestInfo})
	if err != nil {
		return h.apiResponseHandler.ResponseError(c, requestOption, err)
	}

	err = h.options.OnAfter(c, requestOption)
	if err != nil {
		return h.apiResponseHandler.ResponseError(c, requestOption, err)
	}

	return h.apiResponseHandler.ResponseSuccess(c, requestOption, data)
}

func (h *apiHandler[WCtx, RequestInfo, RequestOption]) bodyParserIfRequired(c WCtx, requestOption *RequestOption, requestPtr any) (bool, error) {
	if c.Method() == http.MethodGet {
		return false, nil
	}

	if requestPtr == nil {
		return false, nil
	}

	err := c.BodyParser(requestPtr)
	if err != nil {
		return false, h.apiResponseHandler.ResponseError(c, requestOption, err)
	}

	if IsMultipartForm(c) {
		err = MultipartBodyParser(c, requestPtr)
		if err != nil {
			return false, h.apiResponseHandler.ResponseError(c, requestOption, err)
		}
	}

	return true, nil
}

type Ctx[RequestInfo any] struct {
	Context     context.Context
	RequestInfo *RequestInfo
}

type ApiResponseHandlerOptions[WCtx WebCtx, RequestOption any] struct {
	ResponseSuccess func(c WCtx, requestOption *RequestOption, data any) error
	ResponseError   func(c WCtx, requestOption *RequestOption, err error) error
}

type ApiResponseHandler[WCtx WebCtx, RequestOption any] interface {
	ResponseSuccess(c WCtx, requestOption *RequestOption, data any) error
	ResponseError(c WCtx, requestOption *RequestOption, err error) error
}

type responseHandler[WCtx WebCtx, RequestOption any] struct {
	options *ApiResponseHandlerOptions[WCtx, RequestOption]
}

func (r responseHandler[WCtx, RequestOption]) ResponseError(c WCtx, requestOption *RequestOption, err error) error {
	if r.options.ResponseError != nil {
		return r.options.ResponseError(c, requestOption, err)
	}
	return c.SendString(http.StatusInternalServerError, err.Error())
}

func (r responseHandler[WCtx, RequestOption]) ResponseSuccess(c WCtx, requestOption *RequestOption, data any) error {
	if r.options.ResponseSuccess != nil {
		return r.options.ResponseSuccess(c, requestOption, data)
	}
	c.Status(http.StatusOK)
	return c.JSON(data)
}

func NewApiResponseHandler[WCtx WebCtx, RequestOption any](options *ApiResponseHandlerOptions[WCtx, RequestOption]) ApiResponseHandler[WCtx, RequestOption] {
	return &responseHandler[WCtx, RequestOption]{
		options: options,
	}
}
