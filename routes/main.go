package routes

import (
	"github.com/gin-gonic/gin"
	"hammer-web-api/middleware"
)

//	func Load(router *gin.Engine) {
//		router.Use(gin.Recovery()) // error handle
//
//		router.GET("hello",
//			middleware.CorsMiddleware(),
//			func(ctx *gin.Context) {
//				hello := controllers.HelloController{}
//				hello.Index(ctx)
//			},
//		)
//
//		router.POST("users/add",
//			middleware.AuthMiddleware(),
//			func(ctx *gin.Context) {
//				hello := controllers.UserController{}
//				hello.Add(ctx)
//			},
//		)
//
//		router.POST("auth", func(ctx *gin.Context) {
//			auth := controllers.AuthController{}
//			auth.Index(ctx)
//		})
//	}
func Load(router *gin.Engine) {
	router.Use(gin.Recovery()) // error handle
	router.Use(middleware.CorsMiddleware())

	ApiGroup := router.Group("/api/v1")
	InitUserRouter(ApiGroup)
}