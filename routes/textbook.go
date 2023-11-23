package routes

import (
	"github.com/gin-gonic/gin"
	"hammer-web-api/controllers"
	m "hammer-web-api/middleware"
	"time"
)

func InitTextbookRouter(rg *gin.RouterGroup) {
	textbookRouter := rg.Group("textbooks")
	{
		textbookRouter.POST("", m.AuthMiddleware(), func(c *gin.Context) {
			TextbookCtl := controllers.TextbookController{}
			TextbookCtl.Post(c)
		})

		textbookRouter.GET("", m.AuthMiddleware(), func(c *gin.Context) {
			TextbookCtl := controllers.TextbookController{}
			TextbookCtl.GetUserWorkList(c)
		})

		textbookRouter.GET("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			TextbookCtl := controllers.TextbookController{
				TextbookExpireDuration: time.Hour,
			}
			TextbookCtl.GetUserWorkContent(c)
		})

		textbookRouter.PUT("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			TextbookCtl := controllers.TextbookController{}
			TextbookCtl.Put(c)
		})

		textbookRouter.DELETE("/:id", m.AuthMiddleware(), func(c *gin.Context) {
			TextbookCtl := controllers.TextbookController{}
			TextbookCtl.Delete(c)
		})

		textbookRouter.GET("/subscription", m.AuthMiddleware(), func(c *gin.Context) {
			TextbookCtl := controllers.TextbookController{}
			TextbookCtl.GetSubscription(c)
		})
	}

}
