package controllers

import (
	"LeGinReserve/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type RegStudentController struct{}

func (con RegStudentController) GetRegularStudents(ctx *gin.Context) {
	//	Find data order by add_time from new to old
	studentList := []models.StudentReg{}
	models.DB.Order("add_time desc").Find(&studentList)
	ctx.JSON(http.StatusOK, gin.H{
		"data": studentList,
	})
}

func (con RegStudentController) GetRegularStudent(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal("Error by id(get regular student): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by id(get regular student): " + err.Error(),
		})
		return
	}
	student := models.StudentReg{Id: id}
	models.DB.Find(&student)

	ctx.JSON(http.StatusOK, gin.H{
		"data": student,
	})
}

func (con RegStudentController) DeleteRegularStudent(ctx *gin.Context) {
	//	Get delete student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal("Error by id(delete regular student): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentReg{Id: id}
	models.DB.Find(&student)

	//	Delete student
	models.DB.Delete(&student)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "刪除成功",
	})
}

func (con RegStudentController) ChangeInstallmentStatus(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal("Error by id(change installment status): ", err.Error())
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
	date := student.PayDate.AddDate(0, 1, 0) //	AddDate(year, month, day)

	//	Use Model to update designated student information
	err = models.DB.Model(&student).Updates(models.StudentReg{HavePaid: havePaid, PayDate: date}).Error
	if err != nil {
		log.Fatal("Error by update installment status: ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "繳款成功",
	})
}

func (con RegStudentController) ChangeRegPhysicalCondition(ctx *gin.Context) {
	//	Get chage physical condition student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal("Error by id(change reg physical condition): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentReg{Id: id}
	models.DB.Find(&student)

	physicalCondition := ctx.PostForm("condition")

	//	Update physical condition
	err = models.DB.Model(&student).Update("physical_condition", physicalCondition).Error
	if err != nil {
		log.Fatal("Error by update physical condition: ", err.Error())
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
		log.Fatal("Error by id(buy class): ", err.Error())
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
		log.Fatal("Error by class amount(buy class): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by class amount: " + err.Error(),
		})
		return
	}

	//	Get pay method
	payMethod, err := strconv.Atoi(ctx.PostForm("payMethod"))
	if err != nil {
		log.Fatal("Error by pay method(buy class): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by pay method: " + err.Error(),
		})
		return
	}

	//	Calculate total purchase class amount
	totalClass := student.TotalPurchaseClass + classAmount

	//	Update data by pay method
	err = models.DB.Model(&student).Updates(models.StudentReg{PayMethod: payMethod, TotalPurchaseClass: totalClass, InstallmentAmount: models.Installment[payMethod], HavePaid: 0}).Error
	if err != nil {
		log.Fatal("Error by update total purchase class(buy class):", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "購買資料更新失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "已完成續課",
	})
}
