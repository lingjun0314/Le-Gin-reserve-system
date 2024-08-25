package models

type Student interface {
	GetName()
}

type ReserveStudent struct {
	Id          int
	StudentType int
	StudentId   int
	Student     Student `gorm:"polymorphic:StudentTable"`
}

func (ReserveStudent) TableName() string {
	return "reserve_student"
}
