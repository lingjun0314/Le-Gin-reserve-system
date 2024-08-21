package models

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	//	Read ini file
	config, err := ini.Load("./conf/app.ini")
	if err != nil {
		log.Fatal(err.Error())
		os.Exit(1)
	}

	//	Get value from ini
	ip := config.Section("mysql").Key("ip").String()
	port := config.Section("mysql").Key("port").String()
	user := config.Section("mysql").Key("user").String()
	password := config.Section("mysql").Key("password").String()
	database := config.Section("mysql").Key("database").String()

	//	Set dsn and open database
	dsn := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v", user, password, ip, port, database)
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
}
