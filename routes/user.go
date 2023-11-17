package routes

import (
	"github.com/gin-gonic/gin"
	"hammer-web-api/controllers"
	"hammer-web-api/di"
	m "hammer-web-api/middleware"
)

func InitUserRouter(rg *gin.RouterGroup) {
	userRouter := rg.Group("user")
	di.Zap().Info("set user router group")
	{
		// standard api
		userRouter.POST("", func(c *gin.Context) {
			userCtl := controllers.UserController{}
			userCtl.Post(c)
		})

		userRouter.GET("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			userCtl := controllers.UserController{}
			userCtl.Get(c)
		})

		userRouter.PUT("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			userCtl := controllers.UserController{}
			userCtl.Put(c)
		})

		userRouter.DELETE("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			userCtl := controllers.UserController{}
			userCtl.Delete(c)
		})

		// non-standard api
		userRouter.POST("/login", func(c *gin.Context) {
			userCtl := controllers.UserController{}
			userCtl.Login(c)
		})
		userRouter.GET("/captcha", controllers.GenerateCaptcha)
		userRouter.GET("/sms", controllers.SendSms)
	}
}
