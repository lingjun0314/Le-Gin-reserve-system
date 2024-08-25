package controllers

import (
	"LeGinReserve/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type RegStudentController struct{}

func (con RegStudentController) GetRegularStudents(ctx *gin.Context) {
	//	Find data order by add_time from new to old
	studentList := []models.StudentReg{}
	models.DB.Order("add_time desc").Find(&studentList)

	// Convert payDate to string for each student
	var response []map[string]interface{}
	for _, student := range studentList {
		payDateString := string(student.PayDate)
		studentData := map[string]interface{}{
			"id":                   student.Id,
			"name":                 student.Name,
			"phone":                student.Phone,
			"physical_condition":   student.PhysicalCondition,
			"pay_method":           student.PayMethod,
			"pay_date":             payDateString,
			"installment_amount":   student.InstallmentAmount,
			"have_paid":            student.HavePaid,
			"total_purchase_class": student.TotalPurchaseClass,
			"add_time":             student.AddTime,
		}
		response = append(response, studentData)
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

func (con RegStudentController) GetRegularStudent(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by id(get regular student): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by id(get regular student): " + err.Error(),
		})
		return
	}
	student := models.StudentReg{Id: id}
	models.DB.Find(&student)

	payDateString := string(student.PayDate)

	studentData := map[string]interface{}{
		"id":                   student.Id,
		"name":                 student.Name,
		"phone":                student.Phone,
		"physical_condition":   student.PhysicalCondition,
		"pay_method":           student.PayMethod,
		"pay_date":             payDateString,
		"installment_amount":   student.InstallmentAmount,
		"have_paid":            student.HavePaid,
		"total_purchase_class": student.TotalPurchaseClass,
		"add_time":             student.AddTime,
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": studentData,
	})
}

func (con RegStudentController) CreateRegularStudent(ctx *gin.Context) {
	//	Get student information from user input
	name := ctx.PostForm("name")
	phone := ctx.PostForm("phone")
	physicalCondition := ctx.PostForm("physicalCondition")
	payMethod, err := strconv.Atoi(ctx.PostForm("payMethod"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "輸入資料錯誤，請重試",
		})
		return
	}
	payDate := ctx.PostForm("payDate")
	var payDateBytes []uint8
	if payDate != "" {
		payDateBytes = []uint8(payDate)
	}

	havePaid, err := strconv.Atoi(ctx.PostForm("havePaid"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "輸入資料錯誤，請重試",
		})
		return
	}
	totalPurchaseClass, err := strconv.Atoi(ctx.PostForm("totalPurchaseClass"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "輸入資料錯誤，請重試",
		})
		return
	}

	//	Set information
	student := models.StudentReg{
		Name:               name,
		Phone:              phone,
		PhysicalCondition:  physicalCondition,
		PayMethod:          payMethod,
		PayDate:            payDateBytes,
		InstallmentAmount:  models.Installment[payMethod],
		HavePaid:           havePaid,
		TotalPurchaseClass: totalPurchaseClass,
		AddTime:            time.Now().Unix(),
	}

	//	Create data
	err = models.DB.Create(&student).Error
	if err != nil {
		fmt.Println("Error by create regular student: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "新增資料失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "新增成功",
	})
}

func (con RegStudentController) DeleteRegularStudent(ctx *gin.Context) {
	//	Get delete student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by id(delete regular student): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Delete student
	err = models.DB.Where("id = ?", id).Delete(&models.StudentReg{}).Error
	if err != nil {
		fmt.Println("Error by delete info: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "刪除失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "刪除成功",
	})
}

func (con RegStudentController) ChangeInstallmentStatus(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by id(change installment status): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentReg{Id: id}
	models.DB.Find(&student)

	//	Student pay by full amount
	if student.PayMethod == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "該學生為全額付款",
		})
		return
	}

	//	Student has paid all intallments
	if student.HavePaid+1 > student.InstallmentAmount {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "已經付完所有分期",
		})
		return
	}

	//	Calculate update data
	havePaid := student.HavePaid + 1
	payDateTime, err := time.Parse("2006-01-02", string(student.PayDate))
	if err != nil {
		fmt.Println("Error by parse payDate time: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "轉換時間錯誤",
		})
		return
	}
	date := payDateTime.AddDate(0, 1, 0) //	AddDate(year, month, day)

	//	Use Model to update designated student information
	if student.HavePaid+1 == student.InstallmentAmount {
		err = models.DB.Model(&student).Where("id = ?", id).Select("have_paid", "pay_date").Updates(models.StudentReg{HavePaid: havePaid, PayDate: nil}).Error
		if err != nil {
			fmt.Println("Error by update installment status: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "更新失敗，請重試",
			})
			return
		}
	} else {
		err = models.DB.Model(&student).Where("id = ?", id).Updates(models.StudentReg{HavePaid: havePaid, PayDate: []uint8(date.Format("2006-01-02"))}).Error
		if err != nil {
			fmt.Println("Error by update installment status: ", err.Error())
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"message": "更新失敗，請重試",
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "繳款成功",
	})
}

func (con RegStudentController) ChangeRegPhysicalCondition(ctx *gin.Context) {
	//	Get chage physical condition student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by id(change reg physical condition): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentReg{}

	physicalCondition := ctx.PostForm("condition")

	//	Update physical condition
	err = models.DB.Model(&student).Where("id = ?", id).Update("physical_condition", physicalCondition).Error
	if err != nil {
		fmt.Println("Error by update physical condition: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

func (con RegStudentController) BuyClass(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		fmt.Println("Error by id(buy class): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by id: " + err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentReg{Id: id}
	models.DB.Find(&student)

	//	Installments paid not complete
	if student.PayMethod != 0 && student.HavePaid < student.InstallmentAmount {
		ctx.JSON(http.StatusForbidden, gin.H{
			"message": "學生尚有未繳完的分期，請先完成繳款",
		})
		return
	}

	//	Get purchase class amount
	classAmount, err := strconv.Atoi(ctx.PostForm("classAmount"))
	if err != nil {
		fmt.Println("Error by class amount(buy class): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by class amount: " + err.Error(),
		})
		return
	}

	//	Get pay method
	payMethod, err := strconv.Atoi(ctx.PostForm("payMethod"))
	if err != nil {
		fmt.Println("Error by pay method(buy class): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by pay method: " + err.Error(),
		})
		return
	}

	//	Calculate total purchase class amount
	totalClass := student.TotalPurchaseClass + classAmount

	//	Update data by pay method
	err = models.DB.Model(&models.StudentReg{}).Where("id = ?", id).Updates(models.StudentReg{
		PayMethod:          payMethod,
		TotalPurchaseClass: totalClass,
		InstallmentAmount:  models.Installment[payMethod],
		HavePaid:           1,
		PayDate:            []uint8(time.Now().AddDate(0, 1, 0).Format("2006-01-02")),
	}).Error
	if err != nil {
		fmt.Println("Error by update total purchase class(buy class):", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "購買資料更新失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "已完成續課",
	})
}
