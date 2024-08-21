package controllers

import (
	"LeGinReserve/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ExpStudentController struct{}

func (con ExpStudentController) GetExperienceStudent(ctx *gin.Context) {
	//	Find data order by add_time from new to old
	studentList := []models.StudentExp{}
	models.DB.Order("add_time desc").Find(&studentList)
	ctx.JSON(http.StatusOK, gin.H{
		"data": studentList,
	})
}

func (con ExpStudentController) GetHavePaidDepositExpStudent(ctx *gin.Context) {
	studentList := []models.StudentExp{}
	models.DB.Where("deposit_pay_status = ?", 1).Order("add_time desc").Find(&studentList)
	ctx.JSON(http.StatusOK, gin.H{
		"data": studentList,
	})
}

func (con ExpStudentController) DeleteExperienceStudent(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

	models.DB.Delete(&student)

	ctx.JSON(http.StatusOK, gin.H{
		"message": "刪除成功",
	})
}

func (con ExpStudentController) ChangeExpPhysicalCondition(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	physicalCondition := ctx.PostForm("condition")
	if err != nil {
		log.Fatal(err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	student := models.StudentExp{Id: id}
	models.DB.Find(&student)

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