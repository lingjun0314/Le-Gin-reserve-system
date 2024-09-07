package models

type Holiday struct {
	Id    int
	Year  int
	Month int
	Day   int
}

func (Holiday) TableName() string {
	return "holiday"
}
