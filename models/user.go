package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	gorm.Model
	Username string     `gorm:"type:varchar(255);unique;not null" json:"username" binding:"required"`
	Password string     `gorm:"type:varchar(255);not null" json:"password" binding:"required"`
	Phone    string     `gorm:"type:varchar(11);unique;not null;index" json:"phone" binding:"required"`
	Email    string     `gorm:"type:varchar(255);unique;null" json:"email" binding:"omitempty,email"`
	BirthDay *time.Time `gorm:"type:date;null" json:"birthDay" binding:"omitempty"`
	Profile  string     `gorm:"type:text;null" json:"profile" binding:"omitempty"`
	Avatar   string     `gorm:"type:varchar(255);null" json:"avatar" binding:"omitempty,url"`
}
