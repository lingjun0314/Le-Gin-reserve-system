package controllers

import (
	"LeGinReserve/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ClassController struct{}

func (con ClassController) GetClassByDate(ctx *gin.Context) {
	date, err := time.Parse("2006-01-02", ctx.Param("date"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error date type",
		})
		return
	}

	type clearResult struct {
		Id           int    `json:"id"`
		ReserveDate  string `json:"reserveDate"`
		ReserveTime  string `json:"reserveTime"`
		ClassType    string `json:"classType"`
		ClassEndTime string `json:"classEndTime"`
		ClassRecord  string `json:"classRecord"`
		StudentId    int    `json:"studentId"`
		StudentName  string `json:"studentName"`
		StudentPhone string `json:"studentPhone"`
	}

	reserves := []models.Reserve{}

	models.DB.Where("reserve_date = ?", date.Format("2006-01-02")).Preload("ReserveStudents.StudentExp").Preload("ReserveStudents.StudentReg").Find(&reserves)

	var results []clearResult
	for _, reserve := range reserves {
		if reserve.ReserveStudents.StudentType == "regular" {
			result := clearResult{
				Id:           reserve.Id,
				ReserveDate:  reserve.ReserveDate.Format("2006-01-02"),
				ReserveTime:  reserve.ReserveTime[:len(reserve.ReserveTime)-3],
				ClassType:    reserve.ClassType,
				ClassEndTime: reserve.ClassEndTime[:len(reserve.ClassEndTime)-3],
				ClassRecord:  reserve.ClassRecord,
				StudentId:    reserve.ReserveStudents.StudentReg.Id,
				StudentName:  reserve.ReserveStudents.StudentReg.Name,
				StudentPhone: reserve.ReserveStudents.StudentReg.Phone,
			}
			results = append(results, result)
		} else {
			result := clearResult{
				Id:           reserve.Id,
				ReserveDate:  reserve.ReserveDate.Format("2006-01-02"),
				ReserveTime:  reserve.ReserveTime[:len(reserve.ReserveTime)-3],
				ClassType:    reserve.ClassType,
				ClassEndTime: reserve.ClassEndTime[:len(reserve.ClassEndTime)-3],
				ClassRecord:  reserve.ClassRecord,
				StudentId:    reserve.ReserveStudents.StudentExp.Id,
				StudentName:  reserve.ReserveStudents.StudentExp.Name,
				StudentPhone: reserve.ReserveStudents.StudentExp.Phone,
			}
			results = append(results, result)
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

func (con ClassController) UpdateClassRecord(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error id input",
		})
		return
	}
	text := ctx.PostForm("classRecord")

	reserve := models.Reserve{Id: id}

	models.DB.Find(&reserve)

	reserve.ClassRecord = text
	err = models.DB.Save(&reserve).Error
	if err != nil {
		fmt.Println("Error by update class record: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新失敗，請重試！",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新成功！",
	})
}

func (con ClassController) SetHoliday(ctx *gin.Context) {
	date, err := time.Parse("2006-01-02", ctx.Query("date"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid date format",
		})
		return
	}

	//	Check this date have reserve or not
	reserve := []models.Reserve{}
	models.DB.Where("reserve_date = ?", date.Format("2006-01-02")).Find(&reserve)
	if len(reserve) != 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "該日期已有預約的課程",
		})
		return
	}

	holiday := models.Holiday{
		Year:  date.Year(),
		Month: int(date.Month()),
		Day:   date.Day(),
	}

	err = models.DB.Save(&holiday).Error
	if err != nil {
		fmt.Println("Error by set holiday: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "新增成功！",
	})
}

func (con ClassController) DeleteHoliday(ctx *gin.Context) {
	date, err := time.Parse("2006-01-02", ctx.Query("date"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid date format",
		})
		return
	}

	//	Check this date is holiday or not
	holiday := models.Holiday{}
	if models.DB.Where("year = ? AND month = ? AND day = ?", date.Year(), int(date.Month()), date.Day()).Find(&holiday).RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "此日期不是休假",
		})
		return
	}

	err = models.DB.Delete(&holiday).Error
	if err != nil {
		fmt.Println("Error by delete holiday: ", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "刪除失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "刪除成功",
	})
}

func (con ClassController) GetHolidayByMonth(ctx *gin.Context) {
	year := ctx.Param("year")
	month := ctx.Param("month")

	holidays := []models.Holiday{}

	if models.DB.Where("year = ? AND month = ?", year, month).Find(&holidays).RowsAffected == 0 {
		ctx.JSON(http.StatusNotFound, gin.H{
			"message": "No holiday",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": holidays,
	})
}
