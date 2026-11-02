package ginroutes

import (
	"io"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/proxyadapter/upstream"
	"github.com/gin-gonic/gin"
)

// DefaultProxyDo — имя обработчика в YAML для pass-through на upstream.
const DefaultProxyDo = "auth_client_proxy"

// WriteUpstream пишет ответ upstream без перекодирования конверта.
func WriteUpstream(c *gin.Context, status int, body []byte) {
	if status == 204 || len(body) == 0 {
		c.Status(status)
		return
	}
	c.Data(status, "application/json", body)
}

// NewProxyHandler — проброс method/path/body/query на upstream.
func NewProxyHandler(client *upstream.Client, internalErrCode string) gin.HandlerFunc {
	if internalErrCode == "" {
		internalErrCode = "internal"
	}
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			writeResult(c, ginserver.Internal(internalErrCode, err))
			return
		}
		status, raw, err := client.Forward(
			c.Request.Context(),
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.URL.Query(),
			body,
			c.Request.Header,
		)
		if err != nil {
			writeResult(c, ginserver.Internal(internalErrCode, err))
			return
		}
		WriteUpstream(c, status, raw)
	}
}

func writeResult(c *gin.Context, res ginserver.Result) {
	if res.Status == 204 {
		c.Status(204)
		return
	}
	c.JSON(res.Status, res.Body)
}
