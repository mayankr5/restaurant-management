package routes

import (
	"github.com/mayankr5/v1/restaurant-management/controllers"

	"github.com/gofiber/fiber/v2"
)

func FoodRoutes(app *fiber.App) {
	app.Get("/foods", controllers.GetFoods)
	app.Get("/foods/:food_id", controllers.GetFood)
	app.Post("/foods", controllers.CreateFood)
	app.Patch("/foods/:food_id", controllers.UpdateFood)
}
