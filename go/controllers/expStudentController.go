package controllers

import (
	"LeGinReserve/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ExpStudentController struct{}

func (con ExpStudentController) GetExperienceStudents(ctx *gin.Context) {
	//	Find data order by add_time from new to old
	studentList := []models.StudentExp{}
	models.DB.Order("add_time desc").Find(&studentList)
	ctx.JSON(http.StatusOK, gin.H{
		"data": studentList,
	})
}

func (con ExpStudentController) GetExperienceStudent(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal("Error by id(get experience student): ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "Error by id(get experience student): " + err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

	ctx.JSON(http.StatusOK, gin.H{
		"data": student,
	})
}

func (con ExpStudentController) GetHavePaidDepositExpStudents(ctx *gin.Context) {
	studentList := []models.StudentExp{}
	models.DB.Where("deposit_pay_status = ?", 1).Order("add_time desc").Find(&studentList)
	ctx.JSON(http.StatusOK, gin.H{
		"data": studentList,
	})
}

func (con ExpStudentController) DeleteExperienceStudent(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

	//	Delete student
	models.DB.Delete(&student)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "刪除成功",
	})
}

func (con ExpStudentController) ChangeExpPhysicalCondition(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

	physicalCondition := ctx.PostForm("condition")

	//	Update physical condition
	err = models.DB.Model(&student).Update("physical_condition", physicalCondition).Error
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
	})
}

func (con ExpStudentController) ChangeExpClassPaidStatus(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

	err = models.DB.Model(&student).Update("exp_class_pay_status", !student.ExpClassPayStatus).Error
	if err != nil {
		log.Fatal("Error by update status(chage exp class paid status): ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新狀態失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新狀態成功",
	})
}

func (con ExpStudentController) ChangeDepositStatus(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

	err = models.DB.Model(&student).Update("deposit_pay_status", !student.DepositPayStatus).Error
	if err != nil {
		log.Fatal("Error by update status(chage deposit status): ", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"message": "更新狀態失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新狀態成功",
	})
}

func (con ExpStudentController) ChangeToRegularStudent(ctx *gin.Context) {
	//	Get student id
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	//	Find student from database
	expStudent := models.StudentExp{Id: id}
	models.DB.Find(&expStudent)

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

	//	Set regular student information
	regStudent := models.StudentReg{
		AddTime:            time.Now().Unix(),
		HavePaid:           0,
		InstallmentAmount:  models.Installment[payMethod],
		Name:               expStudent.Name,
		PayDate:            time.Now().AddDate(0, 1, 0),
		PayMethod:          payMethod,
		Phone:              expStudent.Phone,
		PhysicalCondition:  expStudent.PhysicalCondition,
		TotalPurchaseClass: classAmount,
	}

	//	Start a transaction
	tx := models.DB.Begin()

	//	Create regular student
	if err := tx.Create(&regStudent).Error; err != nil {
		tx.Rollback()
		log.Fatal("Error by create reg student: ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "轉換失敗，請重試",
		})
		return
	}

	//	Delete this experience student
	if err := tx.Delete(&expStudent).Error; err != nil {
		tx.Rollback()
		log.Fatal("Error by delete exp student: ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "轉換失敗，請重試",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Fatal("Error by commit transaction: ", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": "轉換失敗，請重試",
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "轉換成功",
	})
}
