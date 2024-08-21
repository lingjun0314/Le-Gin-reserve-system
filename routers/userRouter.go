package routers

import (
	"LeGinReserve/controllers"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(r *gin.Engine) {
	r.GET("/student", controllers.StudentController{}.GetAllStudent)

	r.GET("/student/regular", controllers.RegStudentController{}.GetRegularStudents)
	r.GET("/student/regular/:id", controllers.RegStudentController{}.GetRegularStudent)
	r.DELETE("/student/regular/:id", controllers.RegStudentController{}.DeleteRegularStudent)
	r.PATCH("/student/installment/:id", controllers.RegStudentController{}.ChangeInstallmentStatus)
	r.PUT("/student/regular/physicalCondition/:id", controllers.RegStudentController{}.ChangeRegPhysicalCondition)
	r.PATCH("/student/class/:id", controllers.RegStudentController{}.BuyClass)

	r.GET("/student/experience", controllers.ExpStudentController{}.GetExperienceStudent)
	r.GET("/student/deposit", controllers.ExpStudentController{}.GetHavePaidDepositExpStudent)
	r.DELETE("/student/experience/:id", controllers.ExpStudentController{}.DeleteExperienceStudent)
	r.PUT("/student/experience/physicalCondition/:id", controllers.ExpStudentController{}.ChangeExpPhysicalCondition)

}
