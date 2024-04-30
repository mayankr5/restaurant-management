package routes

import (
	"github.com/mayankr5/v1/restaurant-management/controllers"

	"github.com/gofiber/fiber/v2"
)

func OrderItemRoutes(app *fiber.App) {
	app.Get("/orderItems", controllers.GetOrderItems)
	app.Get("/orderItems/:orderItem_id", controllers.GetOrderItem)
	app.Get("/orderItems-order/:order_id", controllers.GetOrderItemsByOrder)
	app.Post("/orderItems", controllers.CreateOrderItem)
	app.Patch("/orderItems/:orderItem_id", controllers.UpdateOrderItem)
}
