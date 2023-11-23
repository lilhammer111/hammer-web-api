package models

import (
	"errors"
	"gorm.io/gorm"
	"regexp"
)

type Textbook struct {
	gorm.Model
	Title string `gorm:"type:varchar(100);not null;comment: 教程名;uniqueIndex:idx_author_id_title" json:"title,omitempty"`
	Tag   string `gorm:"type:varchar(50);not null;comment: 教程tag" json:"tag,omitempty"`
	Desc  string `gorm:"varchar(255);null;comment: 教程描述" json:"desc,omitempty"`

	AuthorID uint `gorm:"type:int unsigned;not null;uniqueIndex:idx_author_id_title" json:"authorID,omitempty"`
	Author   User `gorm:"foreignKey:AuthorID"`

	CollaboratorID *uint `gorm:"type:int unsigned;null" json:"collaboratorID,omitempty"`
	Collaborator   User  `gorm:"foreignKey:CollaboratorID"`

	IsHot bool `gorm:"not null;default:false" json:"isHot,omitempty"`
	Mark  uint `gorm:"type:tinyint unsigned" json:"mark,omitempty"`
}

type TextbookVersion struct {
	gorm.Model
	No      string `gorm:"type:varchar(20);not null;comment: 版本号" json:"no,omitempty"`
	Content string `gorm:"type:mediumtext;not null;comment: 教程正文" json:"content,omitempty"`

	TextbookID uint `gorm:"int unsigned;not null;index" json:"textbookID,omitempty"`
	Textbook   Textbook
}

func (tv *TextbookVersion) BeforeCreate(tx *gorm.DB) error {
	matched, _ := regexp.MatchString(`^[0-9]+\.[0-9]+\.[0-9]+$`, tv.No)
	var err error
	if !matched {
		err = errors.New("invalid version format")
	}
	return err
}

type UserOperation struct {
	gorm.Model
	UserID     uint `gorm:"type:int unsigned;index;not null" json:"userId,omitempty"`
	User       User
	TextbookID uint `gorm:"int unsigned;index;not null" json:"textbookID,omitempty"`
	Textbook   Textbook
	Operation  uint `gorm:"type:tinyint unsigned;not null;comment:订阅值1,稍后再看值2,已评分值4" json:"operation,omitempty"`
}
