package routes

import (
	"github.com/mayankr5/v1/restaurant-management/controllers"

	"github.com/gofiber/fiber/v2"
)

func MenuRoutes(app *fiber.App) {
	app.Get("/menus", controllers.GetMenus)
	app.Get("/menus/:menu_id", controllers.GetMenu)
	app.Post("/menus", controllers.CreateMenu)
	app.Patch("/menus/:menu_id", controllers.UpdateMenu)
}
