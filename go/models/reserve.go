package models

import "time"

type Reserve struct {
	Id               int       `json:"id"`
	ReserveDate      time.Time `json:"reserve_date"`
	ReserveTime      string    `json:"reserve_time"`
	ReserveStudentId int
	ClassType        int
	ClassEndTime     string
	AddTime          int64
	ClassRecord      string         `json:"class_record"`
	ReserveStudents   ReserveStudent `gorm:"foreignKey:ReserveStudentId"`
}

func (Reserve) TableName() string {
	return "reserve"
}
