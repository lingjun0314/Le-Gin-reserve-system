package routers

import (
	"LeGinReserve/controllers"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(r *gin.Engine) {
	//	All student routers
	r.GET("/student", controllers.StudentController{}.GetAllStudent)

	//	Regular student routers
	r.GET("/student/regular", controllers.RegStudentController{}.GetRegularStudents)
	r.POST("/student/regular", controllers.RegStudentController{}.CreateRegularStudent)
	r.GET("/student/regular/:id", controllers.RegStudentController{}.GetRegularStudent)
	r.DELETE("/student/regular/:id", controllers.RegStudentController{}.DeleteRegularStudent)
	r.PATCH("/student/installment/:id", controllers.RegStudentController{}.ChangeInstallmentStatus)
	r.PUT("/student/regular/physicalCondition/:id", controllers.RegStudentController{}.ChangeRegPhysicalCondition)
	r.PATCH("/student/class/:id", controllers.RegStudentController{}.BuyClass)

	//	Experience student routers
	r.GET("/student/experience", controllers.ExpStudentController{}.GetExperienceStudents)
	r.POST("/student/experience", controllers.ExpStudentController{}.CreateExperienceStudent)
	r.GET("/student/experience/:id", controllers.ExpStudentController{}.GetExperienceStudent)
	r.DELETE("/student/experience/:id", controllers.ExpStudentController{}.DeleteExperienceStudent)
	r.GET("/student/experience/deposit", controllers.ExpStudentController{}.GetHavePaidDepositExpStudents)
	r.PUT("/student/experience/physicalCondition/:id", controllers.ExpStudentController{}.ChangeExpPhysicalCondition)
	r.PATCH("/student/expClassStatus/:id", controllers.ExpStudentController{}.ChangeExpClassPaidStatus)
	r.PATCH("/student/depositStatus/:id", controllers.ExpStudentController{}.ChangeDepositStatus)
	r.POST("/student/experience/regular/:id", controllers.ExpStudentController{}.ChangeToRegularStudent)
	
	//	Reserve routers
	r.GET("/reserve", controllers.ReserveController{}.GetReserveList)
	r.POST("/reserve/regular",controllers.ReserveController{}.CreateRegularReserve)
	r.POST("/reserve/experience",controllers.ReserveController{}.CreateExperienceReserve)
	// r.GET("/reserve/:id", controllers.ReserveController{}.GetReserveDetail)
	r.DELETE("/reserve/:id", controllers.ReserveController{}.DeleteReserve)
	r.GET("/reserve/student/:name",controllers.ReserveController{}.GetReserveByName)
	r.GET("/reserve/free",controllers.ReserveController{}.GetCanReserveTime)
}
