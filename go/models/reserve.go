package models

import "time"

type Reserve struct {
	Id               int       `json:"id"`
	ReserveDate      time.Time `json:"reserveDate"`
	ReserveTime      string    `json:"reserveTime"`
	ReserveStudentId int
	ClassType        string
	ClassEndTime     string
	AddTime          int64
	ClassRecord      string         `json:"classRecord"`
	ReserveStudents  ReserveStudent `gorm:"foreignKey:ReserveStudentId"`
}

func (Reserve) TableName() string {
	return "reserve"
}
