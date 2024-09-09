package models

type StudentExp struct {
	Id                int    `gorm:"primaryKey" json:"id"`
	Name              string `json:"name"`
	Phone             string `json:"phone"`
	PhysicalCondition string `json:"physicalCondition"`
	ExpClassPayStatus bool   `json:"expClassPayStatus"`
	DepositPayStatus  bool   `json:"depositPayStatus"`
	AddTime           int64  `json:"addTime"`
}

func (StudentExp) TableName() string {
	return "student_exp"
}

func (s StudentExp) GetName() string {
	return s.Name
}
