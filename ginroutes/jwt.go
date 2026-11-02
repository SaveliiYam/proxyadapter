package ginroutes

import (
	"strings"

	"github.com/SaveliiYam/ginserver"
	"github.com/SaveliiYam/httpx"
	pajwt "github.com/SaveliiYam/proxyadapter/jwt"
	"github.com/gin-gonic/gin"
)

// RequireAccessJWT — middleware для auth: [jwt] в routes.yaml.
func RequireAccessJWT(parser pajwt.AccessParser, invalidTokenCode string) gin.HandlerFunc {
	if invalidTokenCode == "" {
		invalidTokenCode = "invalid_token"
	}
	return func(c *gin.Context) {
		token, err := bearerToken(c.GetHeader("Authorization"), invalidTokenCode)
		if err != nil {
			ginserver.ErrorEnvelope(c, err)
			c.Abort()
			return
		}
		userUUID, err := parser.ParseAccess(c.Request.Context(), token)
		if err != nil {
			res := ginserver.InvalidToken(invalidTokenCode, err)
			c.JSON(res.Status, res.Body)
			c.Abort()
			return
		}
		c.Set("jwt_user_uuid", userUUID)
		c.Next()
	}
}

func bearerToken(header, errCode string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", httpx.NewError(401, errCode, "invalid or missing access token")
	}
	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", httpx.NewError(401, errCode, "invalid or missing access token")
	}
	return token, nil
}
