package models

import "time"

type StudentReg struct {
	Id                 int
	Name               string
	Phone              int
	PhysicalCondition  string
	PayMethod          int
	PayDate            time.Time
	InstallmentAmount  int
	HavePaid          int
	TotalPurchaseClass int
	AddTime            int64
}

func (StudentReg) TableName() string {
	return "student_reg"
}
