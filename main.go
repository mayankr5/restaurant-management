package main

import (
	"os"

	"github.com/mayankr5/v1/restaurant-management/database"
	"github.com/mayankr5/v1/restaurant-management/middleware"
	"github.com/mayankr5/v1/restaurant-management/routes"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/mongo"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")

func main() {
	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		println("Middleware before routes")
		return c.Next()
	})

	app.Use(middleware.Authentication())

	routes.UserRoutes(app)
	routes.FoodRoutes(app)
	routes.MenuRoutes(app)
	routes.TableRoutes(app)
	routes.OrderRoutes(app)
	routes.OrderItemRoutes(app)
	routes.InvoiceRoutes(app)

	app.Listen(":" + port)
}
