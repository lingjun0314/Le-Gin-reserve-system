package models

type StudentReg struct {
	Id                 int     `gorm:"primaryKey" json:"id"`
	Name               string  `json:"name"`
	Phone              string  `json:"phone"`
	PhysicalCondition  string  `json:"physical_condition"`
	PayMethod          int     `json:"pay_method"`
	PayDate            []uint8 `json:"pay_date"`
	InstallmentAmount  int     `json:"installment_amount"`
	HavePaid           int     `json:"have_paid"`
	TotalPurchaseClass int     `json:"total_purchase_class"`
	HaveReserveClass   int     `json:"have_reserve_class"`
	AddTime            int64   `json:"add_time"`
}

func (StudentReg) TableName() string {
	return "student_reg"
}

func (s StudentReg) GetName() string {
	return s.Name
}
