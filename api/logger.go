package api

import (
	"bytes"
	"exapp-go/pkg/log"
	"exapp-go/pkg/utils"
	"io/ioutil"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		p := c.Request.URL.String()
		req := reqData(c)
		c.Next()

		end := time.Now()
		cost := end.Sub(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()

		resp, _ := c.Get("resp")
		log.Logger().Infow(
			"[GIN]",
			"method", method,
			"url", p,
			"code", statusCode,
			"cost", cost,
			"client_ip", clientIP,
			"req", req,
			"resp", resp,
		)
	}
}

func reqData(c *gin.Context) interface{} {
	var r interface{}
	contentType := c.ContentType()
	if contentType != gin.MIMEJSON {
		return nil
	}
	data, err := c.GetRawData()
	if err != nil {
		return nil
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(data))

	err = utils.Unmarshal(data, &r)
	if err != nil {
		return string(data)
	}

	return r

}
