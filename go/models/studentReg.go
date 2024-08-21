package models

type StudentReg struct {
	Id                 int
	Name               string
	Phone              string
	PhysicalCondition  string
	PayMethod          int
	PayDate            string
	InstallmentAmount  int
	HavePaid          int
	TotalPurchaseClass int
	AddTime            int64
}

func (StudentReg) TableName() string {
	return "student_reg"
}
