package controllers

import (
	"LeGinReserve/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ReserveController struct{}

func (con ReserveController) GetReserveList(ctx *gin.Context) {
	type listResult struct {
		Id        int
		ClassDate []uint8
		ClassTime string
	}

	reserveList := []models.Reserve{}
	models.DB.Find(&reserveList)

	var results []listResult

	for _, reserve := range reserveList {
		result := listResult{
			Id:        reserve.Id,
			ClassDate: reserve.ReserveDate,
			ClassTime: reserve.ReserveTime,
		}
		results = append(results, result)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

func (con ReserveController) GetReserveDetail(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by get id (GetReserve): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by get id (GetReserve): " + err.Error(),
		})
		return
	}

	reserve := models.Reserve{Id: id}

	err = models.DB.Preload("ClassItem").Preload("ReserveStudent.Student").First(&reserve).Error
	if err != nil {
		fmt.Println("Failed  to find reserveby error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed  to find reserveby error: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"reserve_date":  reserve.ReserveDate,
		"reserve_time":  reserve.ReserveTime,
		"class_type":    reserve.ClassType,
		"student_name":  models.GetStudentName(reserve.ReserveStudent.Student),
		"student_phone": models.GetStudentPhone(reserve.ReserveStudent.Student),
	})
}

func (con ReserveController) GetReserveByName(ctx *gin.Context) {
	name := ctx.Param("name")

	reserves := []models.Reserve{}

	//	Join reserve table, reserve_student table and two variety of student table
	err := models.DB.Joins("JOIN reserve_student ON reserve_student.id = reserve.reserve_student_id").
		Joins("LEFT JOIN student_exp ON student_exp.id = reserve_student.student_id AND reserve_student.student_type = ?", 1). //	student_type = 1: Exprience student
		Joins("LEFT JOIN student_reg ON student_reg.id = reserve_student.student_id AND reserve_student.student_type =?", 2).  //	student_type = 2: Regular student
		Where("student_exp.name = ? OR student_reg.name = ?", name, name).
		Preload("ClassItem").Preload("ReserveStudent.Student").
		Find(&reserves).Error

	if err != nil {
		fmt.Println("Error by GetReserveByName: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error by GetReserveByName: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": reserves,
	})
}

func (con ReserveController) DeleteReserve(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by get id (DeleteReserve): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by get id (DeleteReserve): " + err.Error(),
		})
		return
	}

	err = models.DB.Where("id = ?", id).Delete(&models.Reserve{}).Error
	if err != nil {
		fmt.Println("Error while delete reserve: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error while delete reserve: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "刪除成功",
	})
}

func (con ReserveController) GetCanReserveTime(ctx *gin.Context) {
	//	Get all query information
	year, err := strconv.Atoi(ctx.Query("year"))
	if err != nil {
		fmt.Println("Error by year (GetCanReserveTime): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by year (GetCanReserveTime): " + err.Error(),
		})
		return
	}
	month, err := strconv.Atoi(ctx.Query("month"))
	if err != nil {
		fmt.Println("Error by month (GetCanReserveTime): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by month (GetCanReserveTime): " + err.Error(),
		})
		return
	}
	dayType, err := strconv.Atoi(ctx.Query("dayType")) //		type 1: work day	type 2: holiday
	if err != nil {
		fmt.Println("Error by dayType (GetCanReserveTime): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by dayType (GetCanReserveTime): " + err.Error(),
		})
		return
	}
	timeRange, err := strconv.Atoi(ctx.Query("timeRange")) //	range 0: morning	range 1: afternoon	range 2: night
	if err != nil {
		fmt.Println("Error by timeRange (GetCanReserveTime): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by timeRange (GetCanReserveTime): " + err.Error(),
		})
		return
	}
	classType, err := strconv.Atoi(ctx.Query("classType")) //	type 0: regular		type 1: experience
	if err != nil {
		fmt.Println("Error by classType (GetCanReserveTime): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by classType (GetCanReserveTime): " + err.Error(),
		})
		return
	}

	dates := models.GetDateByType(year, month, dayType)

	dateReserve := make(map[time.Time][]models.Reserve)      //	All day reserve records
	morningReserve := make(map[time.Time][]models.Reserve)   //	Morning reserve records
	afternoonReserve := make(map[time.Time][]models.Reserve) //	Afternoon reserve records
	nightReserve := make(map[time.Time][]models.Reserve)     //	Night reserve records

	//	Find all reserves match in dates
	for _, date := range dates {
		reserveList := []models.Reserve{}
		if err := models.DB.Where("reserve_date = ?", date.Format("2006-01-02")).Order("reserve_time ASC").Find(&reserveList).Error; err != nil {
			fmt.Println("Error by find reserveList: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error by find reserveList: " + err.Error(),
			})
			return
		}
		dateReserve[date] = reserveList
	}

	//	Get all free time
	var results []string

	//	Set time range reserve records
	for _, date := range dates {
		if len(dateReserve[date]) == 0 {
			break
		}

		for i := 0; i < len(dateReserve[date]); i++ {
			reserveTime, _ := time.Parse("15:04", dateReserve[date][i].ReserveTime)
			if reserveTime.Before(time.Date(date.Year(), date.Month(), date.Day(), 12, 1, 0, 0, date.Location())) {
				morningReserve[date] = append(morningReserve[date], dateReserve[date][i])
			} else if reserveTime.Before(time.Date(date.Year(), date.Month(), date.Day(), 17, 1, 0, 0, date.Location())) {
				afternoonReserve[date] = append(afternoonReserve[date], dateReserve[date][i])
			} else {
				nightReserve[date] = append(nightReserve[date], dateReserve[date][i])
			}
		}
	}

	switch {
	//	Search morning free time
	case timeRange == 0:
		for _, date := range dates {
			//	If this date has no record
			if len(morningReserve[date]) == 0 {
				results = append(results, fmt.Sprintf("%d/%d/%d 9:00~12:00", date.Year(), date.Month(), date.Day()))
			} else {
				for i := 0; i < len(morningReserve[date]); i++ {
					//	The first record
					if i == 0 {
						reserveTime, _ := time.Parse("15:04", morningReserve[date][i].ReserveTime)
						switch {
						case classType == 0:
							//	If the first class start time to 9:00 has over or equal to 1 hour, append one result
							if reserveTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())).Hours() >= 1 {
								//	Reservations can be made one hour before the next class start
								endTime := reserveTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d 9:00~%s", date.Year(), date.Month(), date.Day(), endTime))
							}
						case classType == 1:
							//	If the first class start time to 9:00 has over or equal to 1.5 hour, append one result
							if reserveTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())).Hours() >= 1.5 {
								//	Reservations can be made one hour before the next class start
								endTime := reserveTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d 9:00~%s", date.Year(), date.Month(), date.Day(), endTime))
							}
						}

					} else if i == len(morningReserve[date])-1 { //	The last record
						// If only two record
						if len(morningReserve[date]) == 2 {
							reserveEndTime, _ := time.Parse("15:04", morningReserve[date][i].ClassEndTime)
							switch {
							case classType == 0:
								//	If the last class end time to 12:00 has over or equal to 1 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location())).Hours() >= 1 {
									startTime := reserveEndTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), startTime))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04", morningReserve[date][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04", morningReserve[date][i].ReserveTime)
								if previousEndTime.Sub(reserveStartTime).Hours() >= 1 {
									endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
								}
							case classType == 1:
								//	If the last class end time to 12:00 has over or equal to 1.5 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									startTime := reserveEndTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), startTime))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04", morningReserve[date][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04", morningReserve[date][i].ReserveTime)
								if previousEndTime.Sub(reserveStartTime).Hours() >= 1.5 {
									endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
								}
							}
						} else {
							reserveEndTime, _ := time.Parse("15:04", morningReserve[date][i].ClassEndTime)
							switch {
							case classType == 0:
								//	If the last class end time to 12:00 has over or equal to 1 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location())).Hours() >= 1 {
									startTime := reserveEndTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), startTime))
								}
							case classType == 1:
								//	If the last class end time to 12:00 has over or equal to 1.5 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									startTime := reserveEndTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), startTime))
								}
							}
						}
					} else {
						switch {
						case classType == 0:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04", morningReserve[date][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04", morningReserve[date][i].ReserveTime)
							if previousEndTime.Sub(reserveStartTime).Hours() >= 1 {
								endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
							}
						case classType == 1:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04", morningReserve[date][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04", morningReserve[date][i].ReserveTime)
							if previousEndTime.Sub(reserveStartTime).Hours() >= 1.5 {
								endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
							}
						}
					}
				}
			}
		}
	//	Search afternoon free time
	case timeRange == 1:
		for _, date := range dates {
			//	If this date has no record
			if len(afternoonReserve[date]) == 0 {
				results = append(results, fmt.Sprintf("%d/%d/%d 13:00~17:00", date.Year(), date.Month(), date.Day()))
			} else {
				for i := 0; i < len(afternoonReserve[date]); i++ {
					//	The first record
					if i == 0 {
						reserveTime, _ := time.Parse("15:04", afternoonReserve[date][i].ReserveTime)
						switch {
						case classType == 0:
							preTimeRangeEndTime, _ := time.Parse("15:04", morningReserve[date][len(morningReserve[date])-1].ClassEndTime)
							if preTimeRangeEndTime.After(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())) {
								if reserveTime.Sub(preTimeRangeEndTime).Hours() >= 1 {
									endTime := reserveTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndTime, endTime))
								}
							} else {
								//	If the first class start time to 13:00 has over or equal to 1 hour, append one result
								if reserveTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())).Hours() >= 1 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 13:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}

						case classType == 1:
							preTimeRangeEndTime, _ := time.Parse("15:04", morningReserve[date][len(morningReserve[date])-1].ClassEndTime)
							if preTimeRangeEndTime.After(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())) {
								//	If the first class start time to 13:00 has over or equal to 1.5 hour, append one result
								if reserveTime.Sub(preTimeRangeEndTime).Hours() >= 1.5 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndTime, endTime))
								}
							} else {
								//	If the first class start time to 13:00 has over or equal to 1.5 hour, append one result
								if reserveTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 13:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}
						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}

					} else if i == len(afternoonReserve[date])-1 { //	The last record
						// If only two record
						if len(afternoonReserve[date]) == 2 {
							reserveEndTime, _ := time.Parse("15:04", afternoonReserve[date][i].ClassEndTime)
							switch {
							case classType == 0:
								//	If the last class end time to 17:00 has over or equal to 1 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location())).Hours() >= 1 {
									startTime := reserveEndTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), startTime))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04", afternoonReserve[date][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04", afternoonReserve[date][i].ReserveTime)
								if previousEndTime.Sub(reserveStartTime).Hours() >= 1 {
									endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
								}
							case classType == 1:
								//	If the last class end time to 17:00 has over or equal to 1.5 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									startTime := reserveEndTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), startTime))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04", afternoonReserve[date][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04", afternoonReserve[date][i].ReserveTime)
								if previousEndTime.Sub(reserveStartTime).Hours() >= 1.5 {
									endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						} else {
							reserveEndTime, _ := time.Parse("15:04", afternoonReserve[date][i].ClassEndTime)
							switch {
							case classType == 0:
								//	If the last class end time to 17:00 has over or equal to 1 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location())).Hours() >= 1 {
									startTime := reserveEndTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), startTime))
								}
							case classType == 1:
								//	If the last class end time to 17:00 has over or equal to 1.5 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									startTime := reserveEndTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), startTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						}
					} else {
						switch {
						case classType == 0:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04", afternoonReserve[date][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04", afternoonReserve[date][i].ReserveTime)
							if previousEndTime.Sub(reserveStartTime).Hours() >= 1 {
								endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
							}
						case classType == 1:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04", afternoonReserve[date][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04", afternoonReserve[date][i].ReserveTime)
							if previousEndTime.Sub(reserveStartTime).Hours() >= 1.5 {
								endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
							}
						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}
					}
				}
			}
		}
	//	Search night free time
	case timeRange == 2:
		for _, date := range dates {
			//	If this date has no record
			if len(nightReserve[date]) == 0 {
				results = append(results, fmt.Sprintf("%d/%d/%d 18:00~21:00", date.Year(), date.Month(), date.Day()))
			} else {
				for i := 0; i < len(nightReserve[date]); i++ {
					//	The first record
					if i == 0 {
						reserveTime, _ := time.Parse("15:04", nightReserve[date][i].ReserveTime)
						switch {
						case classType == 0:
							preTimeRangeEndTime, _ := time.Parse("15:04", afternoonReserve[date][len(morningReserve[date])-1].ClassEndTime)
							if preTimeRangeEndTime.After(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())) {
								//	If the first class start time to 18:00 has over or equal to 1 hour, append one result
								if reserveTime.Sub(preTimeRangeEndTime).Hours() >= 1 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndTime, endTime))
								}
							} else {
								//	If the first class start time to 18:00 has over or equal to 1 hour, append one result
								if reserveTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())).Hours() >= 1 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 18:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}
						case classType == 1:
							preTimeRangeEndTime, _ := time.Parse("15:04", afternoonReserve[date][len(morningReserve[date])-1].ClassEndTime)
							if preTimeRangeEndTime.After(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())) {
								//	If the first class start time to 18:00 has over or equal to 1.5 hour, append one result
								if reserveTime.Sub(preTimeRangeEndTime).Hours() >= 1.5 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndTime, endTime))
								}
							} else {
								//	If the first class start time to 18:00 has over or equal to 1.5 hour, append one result
								if reserveTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 18:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}
						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}

					} else if i == len(nightReserve[date])-1 { //	The last record
						// If only two record
						if len(nightReserve[date]) == 2 {
							reserveEndTime, _ := time.Parse("15:04", nightReserve[date][i].ClassEndTime)
							switch {
							case classType == 0:
								//	If the last class end time to 21:00 has over or equal to 1 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location())).Hours() >= 1 {
									startTime := reserveEndTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~21:00", date.Year(), date.Month(), date.Day(), startTime))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04", nightReserve[date][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04", nightReserve[date][i].ReserveTime)
								if previousEndTime.Sub(reserveStartTime).Hours() >= 1 {
									endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
								}
							case classType == 1:
								//	If the last class end time to 20:30 has over or equal to 1.5 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location())).Hours() >= 1.5 {
									startTime := reserveEndTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~20:30", date.Year(), date.Month(), date.Day(), startTime))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04", nightReserve[date][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04", nightReserve[date][i].ReserveTime)
								if previousEndTime.Sub(reserveStartTime).Hours() >= 1.5 {
									endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						} else {
							reserveEndTime, _ := time.Parse("15:04", nightReserve[date][i].ClassEndTime)
							switch {
							case classType == 0:
								//	If the last class end time to 21:00 has over or equal to 1 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location())).Hours() >= 1 {
									startTime := reserveEndTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~21:00", date.Year(), date.Month(), date.Day(), startTime))
								}
							case classType == 1:
								//	If the last class end time to 20:30 has over or equal to 1.5 hour, append one result
								if reserveEndTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location())).Hours() >= 1.5 {
									startTime := reserveEndTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~20:30", date.Year(), date.Month(), date.Day(), startTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						}
					} else {
						switch {
						case classType == 0:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04", nightReserve[date][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04", nightReserve[date][i].ReserveTime)
							if previousEndTime.Sub(reserveStartTime).Hours() >= 1 {
								endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
							}
						case classType == 1:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04", nightReserve[date][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04", nightReserve[date][i].ReserveTime)
							if previousEndTime.Sub(reserveStartTime).Hours() >= 1.5 {
								endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime, endTime))
							}
						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}
					}
				}
			}
		}
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error time range",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}
