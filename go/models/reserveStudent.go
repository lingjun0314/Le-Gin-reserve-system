package models

type ReserveStudent struct {
	Id          int `json:"id"`
	StudentType int `json:"student_type"`
	StudentId   int
	Student     interface{} `gorm:"polymorphic:StudentTable"`
}

func (ReserveStudent) TableName() string {
	return "reserve_student"
}
