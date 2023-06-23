package handler

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"time"

	logx "github.com/Hyingerrr/mirco-esim/log"

	"github.com/gin-gonic/gin"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		beg := time.Now()
		defer func() {
			if rec := recover(); rec != nil {
				recoverFrom(rec, beg)
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()

		// 写入response Body之前把body拷贝一份到buffer中
		writer := &bodyLogWriter{
			ResponseWriter: c.Writer,
			bodyBuf:        bytes.NewBufferString(""),
		}
		c.Writer = writer
		reqBuf, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatus(http.StatusNotExtended)
			return
		}

		logx.WithFields(logx.Field{
			"cost":   time.Since(beg).Seconds(),
			"code":   c.Writer.Status(),
			"method": c.Request.Method,
			"size":   c.Writer.Size(),
			"host":   c.Request.Host,
			"path":   c.Request.URL.Path,
			"ip":     c.ClientIP(),
			"params": string(reqBuf),
			"err":    c.Errors.ByType(gin.ErrorTypePrivate).String(),
		}).Info("accessLogger")

		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBuf))

		c.Next()

		logx.Infoc(c.Request.Context(), "outerLogger Body: %v, statusCode[%v], cost[%v]", writer.bodyBuf.String(),
			c.Writer.Status(), time.Since(beg).Seconds())
	}
}

type bodyLogWriter struct {
	gin.ResponseWriter
	bodyBuf *bytes.Buffer
}

func (w bodyLogWriter) Write(buf []byte) (int, error) {
	w.bodyBuf.Write(buf)
	return w.ResponseWriter.Write(buf)
}

func recoverFrom(rec interface{}, beg time.Time) {
	var stacktrace string
	for i := 1; i < 4; i++ {
		_, f, l, got := runtime.Caller(i)
		if !got {
			break
		}

		stacktrace += fmt.Sprintf("%s:%d\n", f, l)
	}

	logx.WithFields(logx.Field{
		"cost":  time.Since(beg).Seconds(),
		"stack": stacktrace,
		"err":   rec}).Error("accessLoggerPanic")
}
