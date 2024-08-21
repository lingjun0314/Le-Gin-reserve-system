package controllers

import (
	"LeGinReserve/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StudentController struct{}

func (con StudentController) GetAllStudent(ctx *gin.Context) {
	//	Find data
	regStudentList := []models.StudentReg{}
	expStudentList := []models.StudentExp{}
	models.DB.Order("add_time desc").Find(&regStudentList)
	models.DB.Order("add_time desc").Find(&expStudentList)

	ctx.JSON(http.StatusOK, gin.H{
		"studentReg": regStudentList,
		"studentExp": expStudentList,
	})
}
