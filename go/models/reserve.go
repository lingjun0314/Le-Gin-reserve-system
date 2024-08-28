package models

type Reserve struct {
	Id               int     `json:"id"`
	ReserveDate      []uint8 `json:"reserve_date"`
	ReserveTime      string  `json:"reserve_time"`
	ReserveStudentId int
	ClassType        int
	ClassEndTime     string
	AddTime          int
	ReserveStudent   ReserveStudent `gorm:"foreignKey:ReserveStudentId"`
}

func (Reserve) TableName() string {
	return "reserve"
}
