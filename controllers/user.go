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
	"math/rand"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserController struct {
	BaseModel
	Username string     `gorm:"type:varchar(255);unique;not null" json:"username" binding:"required"`
	Password string     `gorm:"type:varchar(255);not null" json:"password" binding:"required"`
	Phone    string     `gorm:"type:varchar(11);unique;not null;index" json:"phone" binding:"required"`
	Email    string     `gorm:"type:varchar(255);unique;null" json:"email" binding:"omitempty,email"`
	BirthDay *time.Time `gorm:"type:date;null" json:"birth-day" binding:"omitempty"`
	Profile  string     `gorm:"type:text;null" json:"profile" binding:"omitempty"`
	Avatar   string     `gorm:"type:varchar(255);null" json:"avatar" binding:"omitempty,url"`
}

func (t *UserController) TableName() string {
	return "user_controllers"
}

type registerForm struct {
	loginForm
	Username string `json:"username" binding:"required"`
	SmsCode  string `json:"sms-code" binding:"required"`
}

type loginForm struct {
	Phone     string `json:"phone" binding:"required"`
	Password  string `json:"password" binding:"required"`
	CaptchaID string `json:"captcha-id" binding:"required"`
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
			"status":  http.StatusInternalServerError,
			"message": "failed to generate captcha",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "ok",
		"data": gin.H{
			"captcha-id":   id,
			"picture-path": b64s,
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
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	if *response.StatusCode != http.StatusOK {
		di.Zap().Errorf("failed to send sms: %s", *response.Body.Message)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	_, err = di.GoRedis().Set(context.Background(), mobile, smsCode, time.Duration(config.Config.Expire)*time.Second).Result()
	if err != nil {
		di.Zap().Errorf("failed to save the value of %s: %s", smsCode, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
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
			"status":  http.StatusBadRequest,
			"message": "bad request",
		})
		return
	}
	// verify captcha
	// attention that the third param of Verify function is false for testing
	if ok := store.Verify(lf.CaptchaID, lf.Captcha, false); !ok {
		di.Zap().Errorf("wrong captcha: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "wrong captcha",
		})
		return
	}

	// verify password
	if di.Gorm().Where("phone = ?", lf.Phone).First(t).RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"status":  http.StatusNotFound,
			"message": "not yet registered",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(t.Password), []byte(lf.Password))
	if err != nil {
		di.Zap().Errorf("wrong password: %s", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "bad request",
		})
		return
	}

	// generate token and response
	token, err := generateToken(t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "failed to create token",
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       http.StatusOK,
		"message":      "ok",
		"access_token": token,
		"expire_in":    7200,
		"data":         t,
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
			"status":  http.StatusBadRequest,
			"message": "bad request",
		})
		return
	}

	// verify captcha
	if ok := store.Verify(rf.CaptchaID, rf.Captcha, false); !ok {
		di.Zap().Errorf("wrong captcha or captcha id: %s", rf.Captcha)
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  http.StatusBadRequest,
			"message": "wrong captcha",
		})
		return
	}

	// verify sms code
	rightSmsCode, err := di.GoRedis().Get(context.Background(), rf.Phone).Result()
	if err != nil {
		di.Zap().Errorf("failed to get value of %s: %s", rf.Phone, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}

	if rf.SmsCode != rightSmsCode {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "wrong sms code",
		})
		return
	}
	// create user
	hashPWD, err := bcrypt.GenerateFromPassword([]byte(rf.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": err.Error(),
		})
		return
	}
	t.Password = string(hashPWD)

	t.Phone = rf.Phone
	t.Username = rf.Username
	if res := di.Gorm().Save(t); res.RowsAffected == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": res.Error.Error(),
		})
		return
	}

	// generate token and response
	token, err := generateToken(t)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Creation of token fails",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":       http.StatusOK,
		"message":      "ok",
		"access_token": token,
		"expire_in":    7200,
		"data":         t,
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

func generateToken(t *UserController) (string, error) {
	now := time.Now().Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": "http://hammer.wang",                                  // 签发人
		"iat": now,                                                   // 签发时间
		"exp": now + int64(7200),                                     // 过期时间
		"nbf": time.Date(2015, 10, 10, 12, 0, 0, 0, time.UTC).Unix(), // 什么时间之前不可用
		"uid": t.ID,
	})

	return token.SignedString([]byte(xenv.Getenv("HMAC_SECRET").String()))
}
