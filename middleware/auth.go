package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mix-go/xutil/xenv"
	"hammer-web-api/di"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 token
		tokenString := c.GetHeader("Authorization")
		if strings.Index(tokenString, "Bearer ") != 0 {
			di.Zap().Error("failed to get token while parsing tokenString")
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "failed to extract token",
			})
			c.Abort()
			return
		}

		// 解码
		token, err := jwt.Parse(tokenString[7:], func(token *jwt.Token) (interface{}, error) {
			// Don't forget to validate the alg is what you expect:
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(xenv.Getenv("HMAC_SECRET").String()), nil
		})
		if err != nil {
			di.Zap().Error("failed to parse token")
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": err.Error(),
			})
			c.Abort()
			return
		}

		// 保存信息
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Set("payload", claims)
		}

		c.Next()
	}
}
