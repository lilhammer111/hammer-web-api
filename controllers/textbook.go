package controllers

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"gorm.io/gorm"
	"hammer-web-api/di"
	"hammer-web-api/models"
	"net/http"
	"strconv"
	"time"
)

type TextbookController struct {
	TextbookExpireDuration time.Duration
}

func (t *TextbookController) GetSubscription(c *gin.Context) {
	// filter condition for user
	claim, exists := c.Get("payload")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to get jwt payload"})
		return
	}

	payload, ok := claim.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to parse claim"})
	}
	userID := payload["uid"]

	// query table UserOperation
	var subscriptions []models.UserOperation
	res := di.Gorm().Where("user_id = ? AND operation in ?", userID, []string{"1", "3", "5", "7"}).Find(&subscriptions)
	if res.RowsAffected == 0 {
		di.Zap().Errorf("user's subscription is not found: %s", res.Error)
		c.JSON(http.StatusNotFound, gin.H{"message": "user's subscription is not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"data":    subscriptions,
	})
}

func (t *TextbookController) GetUserWorkList(c *gin.Context) {
	// filter condition for user
	userID := parseUserIDFromToken(c)
	if userID == "" {
		return
	}
	// query table Textbook
	var textbooks []models.Textbook
	if res := di.Gorm().Where("author_id = ? ", userID).Find(&textbooks); res.RowsAffected == 0 {
		di.Zap().Errorf("user's work is not found:%s", res.Error)
		c.JSON(http.StatusNotFound, gin.H{
			"message": "user's work is not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"data":    textbooks,
	})
}

func (t *TextbookController) GetUserWorkContent(c *gin.Context) {
	// Get textbook id
	textbookID := c.Param("id")
	tid, err := strconv.ParseUint(textbookID, 10, 0)
	if err != nil {
		di.Zap().Errorf("failed to convert string-type textbook id to uint-type: %s", err)
	}
	// Get user id from token
	userID := parseUserIDFromToken(c)
	if userID == "" {
		return
	}

	// Confirm that it is the resource users self have requested but not other users
	var textbook models.Textbook
	res := di.Gorm().Select("id").Where("id = ? AND author_id = ?", tid, userID).First(&textbook)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"message": "request no permission"})
		} else {
			di.Zap().Errorf("failed to query textbook: %s", res.Error)
			c.JSON(http.StatusInternalServerError, gin.H{"message": "internal server error"})
		}
		return
	}

	// Query Version Table for textbook content but query cache first
	var latestVersion models.TextbookVersion

	content, err := di.GoRedis().Get(context.Background(),
		fmt.Sprintf("id_%d_latest_content", tid),
	).Result()
	if err != nil {
		res = di.Gorm().Where("textbook_id = ?", tid).Order("created_at DESC").First(&latestVersion)
		_, redisErr := di.GoRedis().SetEx(context.Background(),
			fmt.Sprintf("id_%d_latest_content", tid),
			latestVersion.Content,
			t.TextbookExpireDuration).Result()
		if redisErr != nil {
			di.Zap().Errorf("failed to setex %d_latest: %s", tid, redisErr)
		}
	} else {
		res = di.Gorm().Select("no").Where("textbook_id = ?", tid).Order("created_at DESC").First(&latestVersion)
		latestVersion.Content = content
		// TODO: Is it necessary to extend expiration time ?
	}
	if res.RowsAffected == 0 {
		di.Zap().Errorf("failed to get latest version textbook while tid is %d: %s", tid, res.Error)
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("textbook %d not found", tid)})
		return
	}

	// query all versions and version id
	var versions []models.TextbookVersion
	res = di.Gorm().Select("no").Where("textbook_id = ?", tid).Find(&versions)
	if res.RowsAffected == 0 {
		di.Zap().Errorf("failed to get any version of textbook while tid is %d: %s", tid, res.Error)
		c.JSON(http.StatusNotFound, gin.H{"message": fmt.Sprintf("textbook %d not found", tid)})
		return
	}

	// Encapsulate data and response
	allVersionsData := make([]map[string]any, 0)
	for _, v := range versions {
		tempMap := make(map[string]any)
		tempMap["version"] = v.No
		tempMap["vid"] = v.ID
		allVersionsData = append(allVersionsData, tempMap)
	}

	respData := gin.H{
		"latestTextbook": gin.H{
			"content": latestVersion.Content,
			"version": latestVersion.No,
		},
		"allVersions": allVersionsData,
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "OK",
		"data":    respData,
	})

}

func (t *TextbookController) Post(c *gin.Context) {

}

func (t *TextbookController) Put(c *gin.Context) {

}

func (t *TextbookController) Delete(c *gin.Context) {

}

func parseUserIDFromToken(c *gin.Context) any {
	claim, exists := c.Get("payload")
	if !exists {
		di.Zap().Error("payload does not exist")
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to get jwt payload"})
		return ""
	}
	payload, ok := claim.(jwt.MapClaims)
	if !ok {
		di.Zap().Error("payload does not exist")
		c.JSON(http.StatusBadRequest, gin.H{"message": "failed to parse claim"})
		return ""
	}
	uid := payload["uid"]

	return uid
}
