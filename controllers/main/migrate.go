package main

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"hammer-web-api/controllers"
	"log"
)

func main() {
	dsn := "root:root@tcp(127.0.0.1:3306)/hammer_web_api?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{SingularTable: true},
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(&controllers.UserController{})
	if err != nil {
		log.Fatal(err)
	}
}
