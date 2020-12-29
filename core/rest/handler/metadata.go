package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jukylin/esim/pkg/common/meta"
)

func GinMetaDataToCtx() gin.HandlerFunc {
	return func(c *gin.Context) {
		var metadata = new(meta.CommonParams)
		reqBuf, err := c.GetRawData()
		if err != nil {
			c.AbortWithStatus(http.StatusNotExtended)
			return
		}

		err = json.Unmarshal(reqBuf, metadata)
		if err != nil {
			c.AbortWithStatus(http.StatusNotExtended)
			return
		}

		md := meta.MD{
			meta.ProdCd:      metadata.ParseProdCd(),
			meta.AppID:       metadata.AppID,
			meta.MerID:       metadata.ParseMerID(),
			meta.RequestNo:   metadata.RequestNo,
			meta.TranCd:      metadata.ParseTranCd(),
			meta.Method:      c.Request.Method,
			meta.Protocol:    meta.HTTPProtocol,
			meta.Uri:         c.Request.URL.Path,
			meta.ServiceName: "appname",
		}
		rCtx := meta.NewContext(c.Request.Context(), md)
		c.Request = c.Request.WithContext(rCtx)

		// MUST: request body put back to gin context body
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(reqBuf))

		c.Next()
	}
}
