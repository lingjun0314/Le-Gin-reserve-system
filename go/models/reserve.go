package models

type Reserve struct {
	Id             int
	ReserveDate    []uint8
	ReserveTime    string
	ReserveStudent int
	ClassTypeId    int
	AddTime        int
	ClassItem      ClassType `gorm:"foreignKey:ClassTypeId"`
}

func (Reserve) TableName() string {
	return "reserve"
}
