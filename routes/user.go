package routes

import (
	"github.com/gin-gonic/gin"
	"hammer-web-api/controllers"
	m "hammer-web-api/middleware"
)

func InitUserRouter(rg *gin.RouterGroup) {
	userRouter := rg.Group("users")
	{
		// standard api
		userRouter.POST("", func(c *gin.Context) {
			user := controllers.UserController{}
			user.Post(c)
		})

		userRouter.GET("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			user := controllers.UserController{}
			user.Get(c)
		})

		userRouter.PUT("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			user := controllers.UserController{}
			user.Put(c)
		})

		userRouter.DELETE("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			user := controllers.UserController{}
			user.Delete(c)
		})

		// non-standard api
		userRouter.POST("/login", func(c *gin.Context) {
			user := controllers.UserController{}
			user.Login(c)
		})
		userRouter.GET("/captcha", controllers.GenerateCaptcha)
		userRouter.GET("/sms", controllers.SendSms)
	}
}
