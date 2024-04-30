package routes

import (
	"github.com/mayankr5/v1/restaurant-management/controllers"

	"github.com/gofiber/fiber/v2"
)

func InvoiceRoutes(app *fiber.App) {
	app.Get("/invoices", controllers.GetInvoices)
	app.Get("/invoices/:invoice_id", controllers.GetInvoice)
	app.Post("/invoices", controllers.CreateInvoice)
	app.Patch("/invoices/:invoice_id", controllers.UpdateInvoice)
}
