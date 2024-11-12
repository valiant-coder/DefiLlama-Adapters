package api

import (
	"exapp-go/internal/errno"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/rotisserie/eris"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type Response struct {
	Code    uint32                 `json:"code"`
	Message string                 `json:"message"`
	Data    interface{}            `json:"data"`
	Meta    map[string]interface{} `json:"meta"`
}

func success(data interface{}, metaKVs ...interface{}) *Response {
	resp := &Response{
		Data: data,
		Meta: make(map[string]interface{}),
	}
	l := len(metaKVs)
	for i := 0; i < l; i += 2 {
		if key, ok := metaKVs[i].(string); ok && i+1 < l {
			resp.Meta[key] = metaKVs[i+1]
		}
	}
	return resp
}

func fail(code uint32, erMsg string) *Response {
	resp := &Response{
		Code:    code,
		Message: erMsg,
	}
	return resp
}

func Error(c *gin.Context, err error) {
	switch t := err.(type) {
	case validator.ValidationErrors:
		resp := fail(http.StatusBadRequest, err.Error())
		returnJson(c, http.StatusBadRequest, resp)
	case *errno.ParamsError:
		resp := fail(t.Code(), err.Error())
		returnJson(c, http.StatusBadRequest, resp)
	default:
		var targetErr *errno.ParamsError
		if eris.As(err, &targetErr) {
			resp := fail(targetErr.Code(), targetErr.Error())
			returnJson(c, http.StatusBadRequest, resp)
		} else {
			fmt.Printf("unknown error: %v\n", err)
			span := trace.SpanFromContext(c.Request.Context())
			format := eris.NewDefaultStringFormat(eris.FormatOptions{
				InvertOutput: true,
				WithTrace:    true,
				InvertTrace:  true,
				WithExternal: true,
			})
			errString := eris.ToCustomString(err, format)
			span.SetAttributes(attribute.String("unknown_error", errString))
			resp := fail(http.StatusInternalServerError, "service internal error")
			returnJson(c, http.StatusInternalServerError, resp)
		}

	}

}

func OK(c *gin.Context, data interface{}) {
	resp := success(data)
	returnJson(c, http.StatusOK, resp)
}

func List(c *gin.Context, data interface{}, total int64) {
	resp := success(data, "total", total)
	returnJson(c, http.StatusOK, resp)
}

func Created(c *gin.Context, data interface{}) {
	resp := success(data)
	returnJson(c, http.StatusCreated, resp)
}

func NoPermission(c *gin.Context) {
	returnJson(c, http.StatusForbidden, fail(http.StatusForbidden, "no permission"))
	c.Abort()
}

func NoContent(c *gin.Context) {
	returnJson(c, http.StatusNoContent, nil)
}

func Unauthorized(c *gin.Context, msg string) {
	returnJson(c, http.StatusUnauthorized, fail(http.StatusUnauthorized, msg))
}

func SuccessAbort(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}

func returnJson(c *gin.Context, httpCode int, resp *Response) {
	c.Set("resp", resp)
	c.JSON(httpCode, resp)
}
