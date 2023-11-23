package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"hammer-web-api/models"
	"log"
	"os"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/hammer_web_api?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: false},
	})
	if err != nil {
		log.Fatal(err)
	}

	ImportRedisTutorial(db)
	ImportGoTutorial(db)
	ImportVueTutorial(db)
}

func ImportGoTutorial(db *gorm.DB) {
	var err error
	user := models.User{}
	if res := db.First(&user, 1); res.RowsAffected == 0 {
		log.Fatal(err)
	}

	textbook := models.Textbook{
		AuthorID: user.ID,
		Desc:     "go基础教程，主要是golang基础语法",
		Tag:      "golang",
		Title:    "golang基础教程",
	}

	if res := db.Create(&textbook); res.RowsAffected == 0 {
		log.Fatal()
	}

	file, err := os.ReadFile("/home/demon/GolandProjects/hammer-web-api/models/main/go.md")
	if err != nil {
		log.Fatalf("failed to read file go.md:%s", err.Error())
	}

	version := models.TextbookVersion{
		No:         "1.0.0",
		Content:    string(file),
		TextbookID: textbook.ID,
	}
	if res := db.Create(&version); res.RowsAffected == 0 {
		log.Fatal()
	}
}

func ImportRedisTutorial(db *gorm.DB) {
	var err error
	user := models.User{}
	if res := db.First(&user, 1); res.RowsAffected == 0 {
		log.Fatal(err)
	}

	textbook := models.Textbook{
		AuthorID: user.ID,
		Desc:     "redis基础教程，主要是redis基础语法",
		Tag:      "redis",
		Title:    "redis基础教程",
	}

	if res := db.Create(&textbook); res.RowsAffected == 0 {
		log.Fatal()
	}

	file, err := os.ReadFile("/home/demon/GolandProjects/hammer-web-api/models/main/redis-note.md")
	if err != nil {
		log.Fatalf("failed to read file go.md:%s", err.Error())
	}

	version := models.TextbookVersion{
		No:         "1.0.0",
		Content:    string(file),
		TextbookID: textbook.ID,
	}
	if res := db.Create(&version); res.RowsAffected == 0 {
		log.Fatal()
	}
}

func ImportVueTutorial(db *gorm.DB) {
	var err error
	user := models.User{}
	if res := db.First(&user, 1); res.RowsAffected == 0 {
		log.Fatal(err)
	}

	textbook := models.Textbook{
		AuthorID: user.ID,
		Desc:     "vue基础教程，主要是vue基础语法",
		Tag:      "vue",
		Title:    "vue基础教程",
	}

	if res := db.Create(&textbook); res.RowsAffected == 0 {
		log.Fatal()
	}

	file, err := os.ReadFile("/home/demon/GolandProjects/hammer-web-api/models/main/Vue.md")
	if err != nil {
		log.Fatalf("failed to read file go.md:%s", err.Error())
	}

	version := models.TextbookVersion{
		No:         "1.0.0",
		Content:    string(file),
		TextbookID: textbook.ID,
	}
	if res := db.Create(&version); res.RowsAffected == 0 {
		log.Fatal()
	}
}
