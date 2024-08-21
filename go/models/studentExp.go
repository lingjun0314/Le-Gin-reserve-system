package models

type StudentExp struct {
	Id                int
	Name              string
	Phone             int
	PhysicalCondition string
	ExpClassPayStatus bool
	DepositPayStatus  bool
	AddTime           int64
}

func (StudentExp) TableName() string {
	return "student_exp"
}
