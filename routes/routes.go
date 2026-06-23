package routes

import (
	"equipment-borrow-system/controllers"
	"equipment-borrow-system/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	userController := controllers.NewUserController()
	deviceController := controllers.NewDeviceController()
	borrowController := controllers.NewBorrowController()
	statsController := controllers.NewStatsController()

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", userController.Login)
			auth.POST("/register", userController.Register)
		}

		api.Use(middleware.AuthMiddleware())
		{
			user := api.Group("/user")
			{
				user.GET("/profile", userController.GetProfile)
				user.GET("/all", middleware.AdminMiddleware(), userController.GetAllUsers)
			}

			device := api.Group("/device")
			{
				device.GET("", deviceController.GetAll)
				device.GET("/:id", deviceController.GetByID)
				device.POST("", middleware.AdminMiddleware(), deviceController.Create)
				device.PUT("/:id", middleware.AdminMiddleware(), deviceController.Update)
				device.DELETE("/:id", middleware.AdminMiddleware(), deviceController.Delete)
			}

			borrow := api.Group("/borrow")
			{
				borrow.GET("/my", borrowController.GetMyRecords)
				borrow.GET("/all", middleware.AdminMiddleware(), borrowController.GetAllRecords)
				borrow.GET("/overdue", middleware.AdminMiddleware(), borrowController.GetOverdueRecords)
				borrow.GET("/:id", borrowController.GetRecordByID)
				borrow.POST("", borrowController.Borrow)
				borrow.POST("/return", borrowController.Return)
			}

			stats := api.Group("/stats")
			stats.Use(middleware.AdminMiddleware())
			{
				stats.GET("/overview", statsController.GetOverview)
				stats.GET("/category", statsController.GetByCategory)
				stats.GET("/overdue-details", statsController.GetOverdueDetails)
			}
		}
	}
}
