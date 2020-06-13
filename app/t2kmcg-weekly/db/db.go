package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var (
	db *gorm.DB
)

type Config struct {
	Username string
	Password string
	Hostname string
	Port     string
	Database string
}

func Init(c Config) error {
	protocol := "tcp(" + c.Hostname + ":" + c.Port + ")"
	source := c.Username + ":" + c.Password + "@" + protocol + "/" + c.Database

	newDB, err := gorm.Open("mysql", source)
	if err != nil {
		return err
	}

	db = newDB
	return nil
}

func GetDB() *gorm.DB {
	return db
}

func Close() error {
	return db.Close()
}

func AutoMigrate() {
}
