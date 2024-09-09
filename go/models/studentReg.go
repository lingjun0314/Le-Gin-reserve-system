package models

type StudentReg struct {
	Id                 int     `gorm:"primaryKey" json:"id"`
	Name               string  `json:"name"`
	Phone              string  `json:"phone"`
	PhysicalCondition  string  `json:"physicalCondition"`
	PayMethod          int     `json:"payMethod"`
	PayDate            []uint8 `json:"payDate"`
	InstallmentAmount  int     `json:"installmentAmount"`
	HavePaid           int     `json:"havePaid"`
	TotalPurchaseClass int     `json:"totalPurchaseClass"`
	HaveReserveClass   int     `json:"haveReserveClass"`
	AddTime            int64   `json:"addTime"`
}

func (StudentReg) TableName() string {
	return "student_reg"
}

func (s StudentReg) GetName() string {
	return s.Name
}
