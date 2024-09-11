package models

type ReserveStudent struct {
	Id          int    `json:"id"`
	StudentType string `json:"studentType"`
	StudentId   int
	StudentExp  StudentExp `gorm:"foreignKey:StudentId;reference:Id"`
	StudentReg  StudentReg `gorm:"foreignKey:StudentId;reference:Id"`
}

func (ReserveStudent) TableName() string {
	return "reserve_student"
}
