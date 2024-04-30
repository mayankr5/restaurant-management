package routes

import (
	"github.com/mayankr5/v1/restaurant-management/controllers"

	"github.com/gofiber/fiber/v2"
)

func UserRoutes(app *fiber.App) {
	app.Get("/users", controllers.GetUsers)
	app.Get("/users/:user_id", controllers.GetUser)
	app.Post("/users/signup", controllers.SignUp)
	app.Post("/users/login", controllers.Login)
}
