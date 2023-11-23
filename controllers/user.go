package controllers

import (
	"context"
	"fmt"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/mix-go/xutil/xenv"
	bc "github.com/mojocn/base64Captcha"
	"hammer-web-api/config"
	"hammer-web-api/di"
	"hammer-web-api/models"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
}

type registerForm struct {
	loginForm
	Username string `json:"username" binding:"required"`
	SmsCode  string `json:"smsCode" binding:"required"`
}

type loginForm struct {
	Phone     string `json:"phone" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaID string `json:"captchaID" binding:"required"`
	Captcha   string `json:"captcha" binding:"required"`
}

var store = bc.DefaultMemStore

// non-standard api

func GenerateCaptcha(c *gin.Context) {
	cp := bc.NewCaptcha(bc.DefaultDriverDigit, store)
	id, b64s, err := cp.Generate()
	di.Zap().Debugf("captcha id: %s", id)
	if err != nil {
		di.Zap().Errorf("failed to generate captcha: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to generate captcha",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"data": gin.H{
			"captchaID":   id,
			"picturePath": b64s,
		},
	})
}

func SendSms(c *gin.Context) {
	mobile := c.Query("phone")
	accessKeyId := xenv.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID").String()
	accessKeySecret := xenv.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET").String()
	cfg := &openapi.Config{
		// 您的AccessKey ID
		AccessKeyId: &accessKeyId,
		// 您的AccessKey Secret
		AccessKeySecret: &accessKeySecret,
		// 访问的域名
	}
	cfg.Endpoint = tea.String("dysmsapi.aliyuncs.com")

	client, _ := dysmsapi.NewClient(cfg)

	smsCode := generateSmsCode(6)
	//smsCode := "1234"

	request := &dysmsapi.SendSmsRequest{}
	request.SetPhoneNumbers(mobile)
	request.SetSignName("阿里云短信测试")
	request.SetTemplateCode("SMS_154950909")
	request.SetTemplateParam(fmt.Sprintf("{\"code\":%s}", smsCode))

	response, err := client.SendSms(request)
	if err != nil {
		di.Zap().Errorf("failed to send sms: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if *response.StatusCode != http.StatusOK {
		di.Zap().Errorf("failed to send sms: %s", *response.Body.Message)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	_, err = di.GoRedis().Set(context.Background(), mobile, smsCode, time.Duration(config.Config.Expire)*time.Second).Result()
	if err != nil {
		di.Zap().Errorf("failed to save the value of %s: %s", smsCode, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
	})
}

func (t *UserController) Login(c *gin.Context) {
	// bind form
	lf := loginForm{}
	err := c.ShouldBindJSON(&lf)
	if err != nil {
		di.Zap().Errorf("failed to bind form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please check your entry",
		})
		return
	}
	// verify captcha
	// attention that the third param of Verify function is false for testing
	if ok := store.Verify(lf.CaptchaID, lf.Captcha, false); !ok {
		di.Zap().Errorf("wrong captcha: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "wrong captcha",
		})
		return
	}

	// verify password
	user := models.User{}
	if di.Gorm().Where("phone = ?", lf.Phone).First(&user).RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "not yet registered",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(lf.Password))
	if err != nil {
		di.Zap().Errorf("wrong password: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "please check your entry",
		})
		return
	}

	// generate token and response
	token, err := generateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to create token",
		})
	}

	respData := map[string]any{
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"phone":    user.Phone,
		"profile":  user.Profile,
		"birthDay": user.BirthDay,
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "ok",
		"accessToken": token,
		"expireIn":    7200,
		"data":        respData,
	})
}

// standard api

func (t *UserController) Post(c *gin.Context) {
	// bind form
	rf := registerForm{}
	err := c.ShouldBindJSON(&rf)
	if err != nil {
		di.Zap().Errorf("failed to bind form: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "bad request",
		})
		return
	}

	// verify captcha
	if ok := store.Verify(rf.CaptchaID, rf.Captcha, false); !ok {
		di.Zap().Errorf("wrong captcha or captcha id: %s", rf.Captcha)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "wrong captcha",
		})
		return
	}

	// verify sms code
	rightSmsCode, err := di.GoRedis().Get(context.Background(), rf.Phone).Result()
	if err != nil {
		di.Zap().Errorf("failed to get value of %s: %s", rf.Phone, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	if rf.SmsCode != rightSmsCode {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "wrong sms code",
		})
		return
	}
	// create user
	hashPWD, err := bcrypt.GenerateFromPassword([]byte(rf.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": err.Error(),
		})
		return
	}

	user := models.User{}
	user.Password = string(hashPWD)

	user.Phone = rf.Phone
	user.Username = rf.Username
	if res := di.Gorm().Save(&user); res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": res.Error.Error(),
		})
		return
	}

	// generate token and response
	token, err := generateToken(&user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "Creation of token fails",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "ok",
		"access_token": token,
		"expire_in":    7200,
		"data":         user,
	})
}

func (t *UserController) Get(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "ok",
	})
}

func (t *UserController) Put(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "ok",
	})
}

func (t *UserController) Delete(c *gin.Context) {

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "ok",
	})
}

// helper

// generateSmsCode generate a sms code string
func generateSmsCode(width int) string {
	//生成width长度的短信验证码

	numeric := [10]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	r := len(numeric)

	var sb strings.Builder
	for i := 0; i < width; i++ {
		_, _ = fmt.Fprintf(&sb, "%d", numeric[rand.Intn(r)])
	}
	return sb.String()
}

func generateToken(user *models.User) (string, error) {
	now := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "http://hammer.wang",                                  // 签发人
		"iat": now,                                                   // 签发时间
		"exp": now + int64(7200),                                     // 过期时间
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(), // 什么时间之前不可用
		"uid": user.ID,
	})

	return token.SignedString([]byte(xenv.Getenv("HMAC_SECRET").String()))
}
