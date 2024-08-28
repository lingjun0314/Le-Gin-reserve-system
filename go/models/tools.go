package models

import "time"

func GetStudentName(student interface{}) string {
	switch s := student.(type) {
	case StudentExp:
		return s.Name
	case StudentReg:
		return s.Name
	default:
		return "Unknown"
	}
}

func GetStudentPhone(student interface{}) string {
	switch s := student.(type) {
	case StudentExp:
		return s.Phone
	case StudentReg:
		return s.Phone
	default:
		return "Unknown"
	}
}

func GetDateByType(year, month, dayType int) []time.Time {
	var dates []time.Time
	loc := time.Now().Location()
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, loc)
	lastDay := firstDay.AddDate(0, 1, -1)

	if dayType == 1 {
		for date := firstDay; !date.After(lastDay); date = date.AddDate(0, 0, 1) {
			if date.Weekday() != time.Saturday && date.Weekday() != time.Sunday {
				dates = append(dates, date)
			}
		}
	} else if dayType == 2 {
		for date := firstDay; !date.After(lastDay); date = date.AddDate(0, 0, 1) {
			if date.Weekday() == time.Saturday || date.Weekday() == time.Sunday {
				dates = append(dates, date)
			}
		}
	}
	return dates
}
