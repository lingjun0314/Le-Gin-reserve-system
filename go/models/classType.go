package models

type ClassType struct {
	Id       int
	Type     int
	Duration int
}

func (ClassType) TableName() string {
	return "class_type"
}
