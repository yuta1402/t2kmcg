package model

import (
	"time"

	"github.com/jinzhu/gorm"
)

type Contest struct {
	gorm.Model

	Title       string
	TitlePrefix string
	URL         string
	StartTime   time.Time
	EndTime     time.Time
}
