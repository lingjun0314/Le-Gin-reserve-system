package models

type StudentExp struct {
	Id                int    `gorm:"primaryKey" json:"id"`
	Name              string `json:"name"`
	Phone             string `json:"phone"`
	PhysicalCondition string `json:"physical_condition"`
	ExpClassPayStatus bool   `json:"exp_class_pay_status"`
	DepositPayStatus  bool   `json:"deposit_pay_status"`
	AddTime           int64  `json:"add_time"`
}

func (StudentExp) TableName() string {
	return "student_exp"
}
