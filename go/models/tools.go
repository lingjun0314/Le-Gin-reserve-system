package models

import (
	"fmt"
	"time"
)

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

func GetFreeTimeByDate(date time.Time, classType string) []string {
	var results []string
	reserve := []Reserve{}
	DB.Where("reserve_date = ?", date.Format("2006-01-02")).Find(&reserve)

	if len(reserve) == 0 {
		results = append(results, "09:00~21:00")
	} else {
		for i := 0; i < len(reserve); i++ {
			if i == 0 {
				reserveTime, _ := time.Parse("15:04:05", reserve[i].ReserveTime)
				reserveDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveTime.Hour(), reserveTime.Minute(), 0, 0, date.Location())

				switch {
				case classType == "正課":
					if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())).Hours() >= 1 {
						endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
						results = append(results, fmt.Sprintf("9:00~%s", endTime))
					}
				case classType == "體驗課":
					if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())).Hours() >= 1.5 {
						endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
						results = append(results, fmt.Sprintf("9:00~%s", endTime))
					}
				}

				//	If only one record
				if len(reserve) == 1 {
					reserveEndTime, _ := time.Parse("15:04:05", reserve[i].ClassEndTime)
					reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())

					switch {
					case classType == "正課":
						if time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
							results = append(results, fmt.Sprintf("%s~21:00", reserveEndDateTime.Format("15:04")))
						}
					case classType == "體驗課":
						if time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
							results = append(results, fmt.Sprintf("%s~20:30", reserveEndDateTime.Format("15:04")))
						}
					}
				}

			} else if i == len(reserve)-1 { //	The last record
				// If only two record
				if len(reserve) == 2 {
					reserveEndTime, _ := time.Parse("15:04:05", reserve[i].ClassEndTime)
					reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
					switch {
					case classType == "正課":
						previousEndTime, _ := time.Parse("15:04:05", reserve[i-1].ClassEndTime)
						reserveStartTime, _ := time.Parse("15:04:05", reserve[i].ReserveTime)
						if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
							endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
							results = append(results, fmt.Sprintf("%s~%s", previousEndTime.Format("15:04"), endTime))
						}

						if time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
							results = append(results, fmt.Sprintf("%s~21:00", reserveEndDateTime.Format("15:04")))
						}
					case classType == "體驗課":
						if time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
							results = append(results, fmt.Sprintf("%s~20:30", reserveEndDateTime.Format("15:04")))
						}

						previousEndTime, _ := time.Parse("15:04:05", reserve[i-1].ClassEndTime)
						reserveStartTime, _ := time.Parse("15:04:05", reserve[i].ReserveTime)
						if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
							endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
							results = append(results, fmt.Sprintf("%s~%s", previousEndTime.Format("15:04"), endTime))
						}
					}
				} else {
					reserveEndTime, _ := time.Parse("15:04:05", reserve[i].ClassEndTime)
					reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
					switch {
					case classType == "正課":
						if time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
							results = append(results, fmt.Sprintf("%s~21:00", reserveEndDateTime.Format("15:04")))
						}
					case classType == "體驗課":
						if time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
							results = append(results, fmt.Sprintf("%s~20:30", reserveEndDateTime.Format("15:04")))
						}
					}
				}
			} else {
				switch {
				case classType == "正課":
					previousEndTime, _ := time.Parse("15:04:05", reserve[i-1].ClassEndTime)
					reserveStartTime, _ := time.Parse("15:04:05", reserve[i].ReserveTime)
					if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
						endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
						results = append(results, fmt.Sprintf("%s~%s", previousEndTime.Format("15:04"), endTime))
					}
				case classType == "體驗課":
					previousEndTime, _ := time.Parse("15:04:05", reserve[i-1].ClassEndTime)
					reserveStartTime, _ := time.Parse("15:04:05", reserve[i].ReserveTime)
					if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
						endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
						results = append(results, fmt.Sprintf("%s~%s", previousEndTime.Format("15:04"), endTime))
					}
				}
			}
		}

	}
	return results
}
