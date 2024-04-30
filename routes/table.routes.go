package routes

import (
	"github.com/mayankr5/v1/restaurant-management/controllers"

	"github.com/gofiber/fiber/v2"
)

func TableRoutes(app *fiber.App) {
	app.Get("/tables", controllers.GetTables)
	app.Get("/tables/:table_id", controllers.GetTable)
	app.Post("/tables", controllers.CreateTable)
	app.Patch("/tables/:table_id", controllers.UpdateTable)
}
