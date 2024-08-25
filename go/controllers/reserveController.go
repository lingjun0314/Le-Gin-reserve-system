package controllers

import (
	"LeGinReserve/models"
	"net/http"

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
