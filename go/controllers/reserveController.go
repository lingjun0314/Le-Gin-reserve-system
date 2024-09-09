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
		ClassDate string
		ClassTime string
	}

	reserveList := []models.Reserve{}
	models.DB.Find(&reserveList)

	var results []listResult

	for _, reserve := range reserveList {
		result := listResult{
			Id:        reserve.Id,
			ClassDate: reserve.ReserveDate.Format("2006-01-02"),
			ClassTime: reserve.ReserveTime,
		}
		results = append(results, result)
	}
	ctx.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

func (con ReserveController) CreateRegularReserve(ctx *gin.Context) {
	date, err := time.Parse("2006-01-02", ctx.PostForm("date"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "錯誤的日期格式，請重試",
		})
		return
	}
	reserveTime := ctx.PostForm("time")
	studentId, err := strconv.Atoi(ctx.PostForm("studentId"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by student id",
		})
		return
	}

	//	Parse new reserve time and class end time
	newReserveTime, err := time.Parse("15:04", reserveTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid time format",
		})
		return
	}
	newClassEndTime := newReserveTime.Add(time.Hour)

	student := models.StudentReg{Id: studentId}
	models.DB.Find(&student)

	if student.HaveReserveClass+1 > student.TotalPurchaseClass {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "購買的課程數量已經被預約完",
		})
		return
	}

	tx := models.DB.Begin()

	if err = tx.Model(&student).Update("have_reserve_class", student.HaveReserveClass+1).Error; err != nil {
		tx.Rollback()
		fmt.Println("Error by update have_reserve_class: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "預約失敗，請重試",
		})
		return
	}

	//	Find this date reserve records
	reserveList := []models.Reserve{}
	models.DB.Where("reserve_date = ?", date.Format("2006-01-02")).Preload("ReserveStudents").Find(&reserveList)

	//	Find time conflict or not
	for _, reserve := range reserveList {
		existReserveTime, err := time.Parse("15:04:05", reserve.ReserveTime)
		if err != nil {
			fmt.Println("Error parsing existing reserve time: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error parsing existing reserve time",
			})
			return
		}

		existClassEndTime, err := time.Parse("15:04:05", reserve.ClassEndTime)
		if err != nil {
			fmt.Println("Error parsing existing class end time: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error parsing existing class end time",
			})
			return
		}

		//	Judge time conflict or not
		if newReserveTime.Before(existClassEndTime) && newClassEndTime.After(existReserveTime) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "課程時間衝突，請重新選擇課程時間",
			})
			return
		}
	}

	//	Create reserve student record
	reserveStudent := models.ReserveStudent{
		StudentType: "regular",
		StudentId:   studentId,
	}
	err = tx.Create(&reserveStudent).Error
	if err != nil {
		tx.Rollback()
		fmt.Println("Error by create reserve student: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	// Create reserve record
	newReserve := models.Reserve{
		ReserveDate:      date,
		ReserveTime:      reserveTime,
		ReserveStudentId: reserveStudent.Id,
		ClassType:        "正課",
		ClassEndTime:     newClassEndTime.Format("15:04:05"),
		ClassRecord:      "",
		AddTime:          time.Now().Unix(),
	}

	err = tx.Create(&newReserve).Error
	if err != nil {
		tx.Rollback()
		fmt.Println("Error by create reserve: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fmt.Println("Error by commit transaction: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "新增成功",
	})
}

func (con ReserveController) CreateExperienceReserve(ctx *gin.Context) {
	date, err := time.Parse("2006-01-02", ctx.PostForm("date"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "錯誤的日期格式，請重試",
		})
		return
	}
	reserveTime := ctx.PostForm("time")
	studentName := ctx.PostForm("name")
	studentPhone := ctx.PostForm("phone")

	//	Parse new reserve time and class end time
	newReserveTime, err := time.Parse("15:04", reserveTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid time format",
		})
		return
	}
	newClassEndTime := newReserveTime.Add(time.Hour + 30*time.Minute)

	allExpStudent := []models.StudentExp{}
	models.DB.Find(&allExpStudent)

	//	Check this student has already reserved class or not
	for _, student := range allExpStudent {
		if student.Name == studentName {
			ctx.JSON(http.StatusBadRequest, gin.H{
				"message": "該學生已經預約過體驗課",
			})
			return
		}
	}

	//	Find this date reserve records
	reserveList := []models.Reserve{}
	models.DB.Where("reserve_date = ?", date.Format("2006-01-02")).Find(&reserveList)

	//	Find time conflict or not
	for _, reserve := range reserveList {
		existReserveTime, err := time.Parse("15:04:05", reserve.ReserveTime)
		if err != nil {
			fmt.Println("Error parsing existing reserve time: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error parsing existing reserve time",
			})
			return
		}

		existClassEndTime, err := time.Parse("15:04:05", reserve.ClassEndTime)
		if err != nil {
			fmt.Println("Error parsing existing class end time: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error parsing existing class end time",
			})
			return
		}

		//	Judge time conflict or not
		if newReserveTime.Before(existClassEndTime) && newClassEndTime.After(existReserveTime) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "課程時間衝突，請重新選擇課程時間",
			})
			return
		}
	}

	tx := models.DB.Begin()

	//	Create experience student data
	expStudent := models.StudentExp{
		Name:              studentName,
		Phone:             studentPhone,
		PhysicalCondition: "",
		ExpClassPayStatus: false,
		DepositPayStatus:  false,
		AddTime:           time.Now().Unix(),
	}

	err = tx.Create(&expStudent).Error
	if err != nil {
		tx.Rollback()
		fmt.Println("Error while create experience student in reserve: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	//	Create reserve student record
	reserveStudent := models.ReserveStudent{
		StudentType: "experience",
		StudentId:   expStudent.Id,
	}
	err = tx.Create(&reserveStudent).Error
	if err != nil {
		tx.Rollback()
		fmt.Println("Error by create reserve student: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	// Create reserve record
	newReserve := models.Reserve{
		ReserveDate:      date,
		ReserveTime:      reserveTime,
		ReserveStudentId: reserveStudent.Id,
		ClassType:        "體驗課",
		ClassEndTime:     newClassEndTime.Format("15:04:05"),
		ClassRecord:      "",
		AddTime:          time.Now().Unix(),
	}

	err = tx.Create(&newReserve).Error
	if err != nil {
		tx.Rollback()
		fmt.Println("Error by create reserve: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		fmt.Println("Error by commit transaction: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "新增成功",
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

	err = models.DB.Preload("ReserveStudents.StudentExp").Preload("ReserveStudents.StudentReg").First(&reserve).Error
	if err != nil {
		fmt.Println("Failed  to find reserveby error: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Failed  to find reserveby error: " + err.Error(),
		})
		return
	}

	if reserve.ReserveStudents.StudentType == "experience" {
		ctx.JSON(http.StatusOK, gin.H{
			"reserveId":    reserve.Id,
			"reserveDate":  reserve.ReserveDate.Format("2006-01-02"),
			"reserveTime":  reserve.ReserveTime[:len(reserve.ReserveTime)-3],
			"classType":    reserve.ClassType,
			"studentId":    reserve.ReserveStudents.StudentId,
			"studentName":  reserve.ReserveStudents.StudentExp.Name,
			"studentPhone": reserve.ReserveStudents.StudentExp.Phone,
			"classRecord":  reserve.ClassRecord,
		})
	} else if reserve.ReserveStudents.StudentType == "regular" {
		ctx.JSON(http.StatusOK, gin.H{
			"reserveId":    reserve.Id,
			"reserveDate":  reserve.ReserveDate.Format("2006-01-02"),
			"reserveTime":  reserve.ReserveTime[:len(reserve.ReserveTime)-3],
			"classType":    reserve.ClassType,
			"studentId":    reserve.ReserveStudents.StudentId,
			"studentName":  reserve.ReserveStudents.StudentReg.Name,
			"studentPhone": reserve.ReserveStudents.StudentReg.Phone,
			"classRecord":  reserve.ClassRecord,
		})
	} else {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error student type",
		})
		return
	}

}

func (con ReserveController) GetReserveByName(ctx *gin.Context) {
	name := ctx.Param("name")

	reserves := []models.Reserve{}

	//	Join reserve table, reserve_student table and two variety of student table
	err := models.DB.Joins("JOIN reserve_student ON reserve_student.id = reserve.reserve_student_id").
		Joins("LEFT JOIN student_exp ON student_exp.id = reserve_student.student_id AND reserve_student.student_type = ?", 1). //	student_type = 1: Exprience student
		Joins("LEFT JOIN student_reg ON student_reg.id = reserve_student.student_id AND reserve_student.student_type = ?", 2). //	student_type = 2: Regular student
		Where("student_exp.name = ? OR student_reg.name = ?", name, name).
		Preload("ReserveStudents").
		Find(&reserves).Error

	if err != nil {
		fmt.Println("Error by GetReserveByName: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "Error by GetReserveByName: " + err.Error(),
		})
		return
	}

	//	Clearify result
	type clearResult struct {
		Id               int
		ReserveDate      string
		ReserveTime      string
		ReserveStudentId int
		ClassType        string
		ClassEndTime     string
		ClassRecord      string
	}
	var results []clearResult

	for _, reserve := range reserves {
		result := clearResult{
			Id:               reserve.Id,
			ReserveDate:      reserve.ReserveDate.Format("2006-01-02"),
			ReserveTime:      reserve.ReserveTime[:len(reserve.ReserveTime)-3],
			ReserveStudentId: reserve.ReserveStudentId,
			ClassType:        reserve.ClassType,
			ClassEndTime:     reserve.ClassEndTime,
			ClassRecord:      reserve.ClassRecord,
		}
		results = append(results, result)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": results,
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

func (con ReserveController) UpdateReserveData(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by get id (DeleteReserve): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by get id (DeleteReserve): " + err.Error(),
		})
		return
	}
	date, err := time.Parse("2006-01-02", ctx.PostForm("date"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error date format",
		})
		return
	}
	newReserveTime, err := time.Parse("15:04", ctx.PostForm("time"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error time format",
		})
		return
	}
	classType := ctx.PostForm("classType")

	var newClassEndTime time.Time
	if classType == "正課" {
		newClassEndTime = newReserveTime.Add(time.Hour)
	} else if classType == "體驗課" {
		newClassEndTime = newReserveTime.Add(1*time.Hour + 30*time.Minute)
	}

	//	Find this date reserve records
	reserveList := []models.Reserve{}
	models.DB.Where("reserve_date = ?", date.Format("2006-01-02")).Preload("ReserveStudents").Find(&reserveList)

	//	Find time conflict or not
	for _, reserve := range reserveList {
		existReserveTime, err := time.Parse("15:04:05", reserve.ReserveTime)
		if err != nil {
			fmt.Println("Error parsing existing reserve time: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error parsing existing reserve time",
			})
			return
		}

		existClassEndTime, err := time.Parse("15:04:05", reserve.ClassEndTime)
		if err != nil {
			fmt.Println("Error parsing existing class end time: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error parsing existing class end time",
			})
			return
		}

		//	Judge time conflict or not
		if newReserveTime.Before(existClassEndTime) && newClassEndTime.After(existReserveTime) {
			ctx.JSON(http.StatusConflict, gin.H{
				"message": "課程時間衝突，請重新選擇課程時間",
			})
			return
		}
	}

	reserve := models.Reserve{Id: id}
	models.DB.First(&reserve)

	reserve.ReserveDate = date
	reserve.ReserveTime = newReserveTime.Format("15:04:05")
	models.DB.Save(&reserve)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新預約時間成功！",
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

	dateReserve := make(map[string][]models.Reserve)      //	All day reserve records
	morningReserve := make(map[string][]models.Reserve)   //	Morning reserve records
	afternoonReserve := make(map[string][]models.Reserve) //	Afternoon reserve records
	nightReserve := make(map[string][]models.Reserve)     //	Night reserve records

	holidays := []models.Holiday{}
	models.DB.Where("year = ? AND month = ?", year, month).Find(&holidays)

	//	Find all reserves match in dates
	for _, date := range dates {
		//	Except holiday
		isHoliday := false
		for _, holiday := range holidays {
			if date.Day() == holiday.Day {
				isHoliday = true
				break
			}
		}
		if isHoliday {
			continue
		}

		reserveList := []models.Reserve{}
		if err := models.DB.Where("reserve_date = ?", date).Order("reserve_time ASC").Find(&reserveList).Error; err != nil {
			fmt.Println("Error by find reserveList: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "Error by find reserveList: " + err.Error(),
			})
			return
		}
		dateReserve[date.Format("2006-01-02")] = reserveList
	}

	//	Get all free time
	var results []string

	//	Set time range reserve records
	for _, date := range dates {
		dateStr := date.Format("2006-01-02")
		if len(dateReserve[date.Format("2006-01-02")]) == 0 {
			continue
		}
		for i := 0; i < len(dateReserve[date.Format("2006-01-02")]); i++ {
			reserveTime, _ := time.Parse("15:04:05", dateReserve[dateStr][i].ReserveTime)
			reserveDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveTime.Hour(), reserveTime.Minute(), 0, 0, date.Location())
			if reserveDateTime.Before(time.Date(date.Year(), date.Month(), date.Day(), 12, 1, 0, 0, date.Location())) {
				morningReserve[dateStr] = append(morningReserve[dateStr], dateReserve[dateStr][i])
			} else if reserveDateTime.Before(time.Date(date.Year(), date.Month(), date.Day(), 17, 1, 0, 0, date.Location())) {
				afternoonReserve[dateStr] = append(afternoonReserve[dateStr], dateReserve[dateStr][i])
			} else {
				nightReserve[dateStr] = append(nightReserve[dateStr], dateReserve[dateStr][i])
			}
		}
	}

	switch {
	//	Search morning free time
	case timeRange == 0:
		for _, date := range dates {
			if date.Before(time.Now()) {
				continue
			}
			dateStr := date.Format("2006-01-02")
			//	If this date has no record
			if len(morningReserve[dateStr]) == 0 {
				results = append(results, fmt.Sprintf("%d/%d/%d 9:00~12:00", date.Year(), date.Month(), date.Day()))
			} else {
				for i := 0; i < len(morningReserve[dateStr]); i++ {
					//	The first record
					if i == 0 {
						reserveTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ReserveTime)
						reserveDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveTime.Hour(), reserveTime.Minute(), 0, 0, date.Location())

						switch {
						case classType == 0:
							//	If the first class start time to 9:00 has over or equal to 1 hour, append one result
							if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())).Hours() >= 1 {
								//	Reservations can be made one hour before the next class start
								endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d 9:00~%s", date.Year(), date.Month(), date.Day(), endTime))
							}
						case classType == 1:
							//	If the first class start time to 9:00 has over or equal to 1.5 hour, append one result
							if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 9, 0, 0, 0, date.Location())).Hours() >= 1.5 {
								//	Reservations can be made one hour before the next class start
								endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d 9:00~%s", date.Year(), date.Month(), date.Day(), endTime))
							}
						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}

						//	If only one record
						if len(morningReserve[dateStr]) == 1 {
							reserveEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())

							switch {
							case classType == 0:
								//	If the last class end time to 12:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 12:00 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						}

					} else if i == len(morningReserve[dateStr])-1 { //	The last record
						// If only two record
						if len(morningReserve[dateStr]) == 2 {
							reserveEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ReserveTime)
								if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
									endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
								}

								//	If the last class end time to 12:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 12:00 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ReserveTime)
								if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
									endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						} else {
							reserveEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								//	If the last class end time to 12:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 12:00 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 12, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~12:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
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
							previousEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ReserveTime)
							if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
								endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
							}
						case classType == 1:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04:05", morningReserve[dateStr][i].ReserveTime)
							if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
								endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
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
	//	Search afternoon free time
	case timeRange == 1:
		for _, date := range dates {
			if date.Before(time.Now()) {
				continue
			}
			dateStr := date.Format("2006-01-02")
			//	If this date has no record
			if len(afternoonReserve[dateStr]) == 0 {
				results = append(results, fmt.Sprintf("%d/%d/%d 13:00~17:00", date.Year(), date.Month(), date.Day()))
			} else {
				for i := 0; i < len(afternoonReserve[dateStr]); i++ {
					//	The first record
					if i == 0 {
						reserveTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ReserveTime)
						reserveDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveTime.Hour(), reserveTime.Minute(), 0, 0, date.Location())

						switch {
						case classType == 0:
							//	If the class end time of previous time range exceeds the start time of the afternoon period
							if len(morningReserve[dateStr]) != 0 {
								preTimeRangeEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][len(morningReserve[dateStr])-1].ClassEndTime)
								preTimeRangeEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), preTimeRangeEndTime.Hour(), preTimeRangeEndTime.Minute(), 0, 0, date.Location())
								if preTimeRangeEndDateTime.After(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())) {
									//	If the first class start time to 13:00 has over or equal to 1 hour, append one result
									if reserveDateTime.Sub(preTimeRangeEndDateTime).Hours() >= 1 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndDateTime.Format("15:04"), endTime))
									}
								} else {
									//	If the first class start time to 13:00 has over or equal to 1 hour, append one result
									if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())).Hours() >= 1 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d 13:00~%s", date.Year(), date.Month(), date.Day(), endTime))
									}
								}
							} else {
								//	If the first class start time to 13:00 has over or equal to 1 hour, append one result
								if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())).Hours() >= 1 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 13:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}

						case classType == 1:
							if len(morningReserve[dateStr]) != 0 {
								preTimeRangeEndTime, _ := time.Parse("15:04:05", morningReserve[dateStr][len(morningReserve[dateStr])-1].ClassEndTime)
								preTimeRangeEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), preTimeRangeEndTime.Hour(), preTimeRangeEndTime.Minute(), 0, 0, date.Location())
								if preTimeRangeEndDateTime.After(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())) {
									//	If the first class start time to 13:00 has over or equal to 1.5 hour, append one result
									if reserveDateTime.Sub(preTimeRangeEndDateTime).Hours() >= 1.5 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndTime.Format("15:04"), endTime))
									}
								} else {
									//	If the first class start time to 13:00 has over or equal to 1.5 hour, append one result
									if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())).Hours() >= 1.5 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d 13:00~%s", date.Year(), date.Month(), date.Day(), endTime))
									}
								}
							} else {
								//	If the first class start time to 18:00 has over or equal to 1.5 hour, append one result
								if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 13, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 13:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}

						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}

						//	If only one record
						if len(afternoonReserve[dateStr]) == 1 {
							reserveEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								//	If the last class end time to 17:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 17:00 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 17, 00, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						}

					} else if i == len(afternoonReserve[dateStr])-1 { //	The last record
						// If only two record
						if len(afternoonReserve[dateStr]) == 2 {
							reserveEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ReserveTime)
								if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
									endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
								}

								//	If the last class end time to 17:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 17:00 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 17, 00, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ReserveTime)
								if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
									endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						} else {
							reserveEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								//	If the last class end time to 17:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 17, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 17:00 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 17, 00, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~17:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
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
							previousEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ReserveTime)
							if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
								endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
							}
						case classType == 1:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][i].ReserveTime)
							if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
								endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
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
			if date.Before(time.Now()) {
				continue
			}
			dateStr := date.Format("2006-01-02")
			//	If this date has no record
			if len(nightReserve[dateStr]) == 0 {
				results = append(results, fmt.Sprintf("%d/%d/%d 18:00~21:00", date.Year(), date.Month(), date.Day()))
			} else {
				for i := 0; i < len(nightReserve[dateStr]); i++ {
					//	The first record
					if i == 0 {
						reserveTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ReserveTime)
						reserveDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveTime.Hour(), reserveTime.Minute(), 0, 0, date.Location())

						switch {
						case classType == 0:
							//	If the class end time of previous time range exceeds the start time of the night period
							if len(afternoonReserve[dateStr]) != 0 {
								preTimeRangeEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][len(afternoonReserve[dateStr])-1].ClassEndTime)
								preTimeRangeEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), preTimeRangeEndTime.Hour(), preTimeRangeEndTime.Minute(), 0, 0, date.Location())
								if preTimeRangeEndDateTime.After(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())) {
									//	If the first class start time to 18:00 has over or equal to 1 hour, append one result
									if reserveDateTime.Sub(preTimeRangeEndDateTime).Hours() >= 1 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndDateTime.Format("15:04"), endTime))
									}
								} else {
									//	If the first class start time to 18:00 has over or equal to 1 hour, append one result
									if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())).Hours() >= 1 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d 18:00~%s", date.Year(), date.Month(), date.Day(), endTime))
									}
								}
							} else {
								//	If the first class start time to 18:00 has over or equal to 1 hour, append one result
								if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())).Hours() >= 1 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveDateTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 18:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}

						case classType == 1:
							if len(afternoonReserve[dateStr]) != 0 {
								preTimeRangeEndTime, _ := time.Parse("15:04:05", afternoonReserve[dateStr][len(afternoonReserve[dateStr])-1].ClassEndTime)
								preTimeRangeEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), preTimeRangeEndTime.Hour(), preTimeRangeEndTime.Minute(), 0, 0, date.Location())
								if preTimeRangeEndDateTime.After(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())) {
									//	If the first class start time to 18:00 has over or equal to 1.5 hour, append one result
									if reserveDateTime.Sub(preTimeRangeEndDateTime).Hours() >= 1.5 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), preTimeRangeEndTime.Format("15:04"), endTime))
									}
								} else {
									//	If the first class start time to 18:00 has over or equal to 1.5 hour, append one result
									if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())).Hours() >= 1.5 {
										//	Reservations can be made one hour before the next class start
										endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
										results = append(results, fmt.Sprintf("%d/%d/%d 18:00~%s", date.Year(), date.Month(), date.Day(), endTime))
									}
								}
							} else {
								//	If the first class start time to 18:00 has over or equal to 1.5 hour, append one result
								if reserveDateTime.Sub(time.Date(date.Year(), date.Month(), date.Day(), 18, 0, 0, 0, date.Location())).Hours() >= 1.5 {
									//	Reservations can be made one hour before the next class start
									endTime := reserveDateTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d 18:00~%s", date.Year(), date.Month(), date.Day(), endTime))
								}
							}

						default:
							ctx.JSON(http.StatusBadRequest, gin.H{
								"message": "Error class type",
							})
							return
						}

						//	If only one record
						if len(nightReserve[dateStr]) == 1 {
							reserveEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								//	If the last class end time to 21:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~21:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 20:30 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~20:30", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						}

					} else if i == len(nightReserve[dateStr])-1 { //	The last record
						// If only two record
						if len(nightReserve[dateStr]) == 2 {
							reserveEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ReserveTime)
								if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
									endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
								}

								//	If the last class end time to 21:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~21:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 20:30 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~20:30", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}

								// Calculate time during two record
								previousEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i-1].ClassEndTime)
								reserveStartTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ReserveTime)
								if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
									endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
									results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
								}
							default:
								ctx.JSON(http.StatusBadRequest, gin.H{
									"message": "Error class type",
								})
								return
							}
						} else {
							reserveEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ClassEndTime)
							reserveEndDateTime := time.Date(date.Year(), date.Month(), date.Day(), reserveEndTime.Hour(), reserveEndTime.Minute(), 0, 0, date.Location())
							switch {
							case classType == 0:
								//	If the last class end time to 21:00 has over or equal to 1 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 21, 0, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~21:00", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
								}
							case classType == 1:
								//	If the last class end time to 20:30 has over or equal to 1.5 hour, append one result
								if time.Date(date.Year(), date.Month(), date.Day(), 20, 30, 0, 0, date.Location()).Sub(reserveEndDateTime).Hours() >= 1.5 {
									results = append(results, fmt.Sprintf("%d/%d/%d %s~20:30", date.Year(), date.Month(), date.Day(), reserveEndDateTime.Format("15:04")))
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
							previousEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ReserveTime)
							if reserveStartTime.Sub(previousEndTime).Hours() >= 1 {
								endTime := reserveStartTime.Add(-1 * time.Hour).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
							}
						case classType == 1:
							// Calculate time during two record
							previousEndTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i-1].ClassEndTime)
							reserveStartTime, _ := time.Parse("15:04:05", nightReserve[dateStr][i].ReserveTime)
							if reserveStartTime.Sub(previousEndTime).Hours() >= 1.5 {
								endTime := reserveStartTime.Add(-1*time.Hour - 30*time.Minute).Format("15:04")
								results = append(results, fmt.Sprintf("%d/%d/%d %s~%s", date.Year(), date.Month(), date.Day(), previousEndTime.Format("15:04"), endTime))
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
