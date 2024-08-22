package models

import (
	"fmt"
	"os"
	"sync"

	"gopkg.in/ini.v1"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var Installment map[int]int
var once sync.Once

func init() {
	//	Init installment
	Installment = make(map[int]int)

	//	Read ini file
	config, err := ini.Load("./conf/app.ini")
	if err != nil {
		fmt.Println(err.Error())
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
		fmt.Println(err)
	}

	once.Do(func() {
		Installment[0] = 0
		Installment[1] = 3
		Installment[2] = 4
		Installment[3] = 5
		Installment[4] = 6
		Installment[5] = 7
		Installment[6] = 8
		Installment[7] = 9
		Installment[8] = 10
	})
}
