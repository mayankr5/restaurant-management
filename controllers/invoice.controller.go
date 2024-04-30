package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mayankr5/v1/restaurant-management/database"
	"github.com/mayankr5/v1/restaurant-management/models"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type InvoiceViewFormat struct {
	Invoice_id       string
	Payment_method   string
	Order_id         string
	Payment_status   *string
	Payment_due      interface{}
	Table_number     interface{}
	Payment_due_date time.Time
	Order_details    interface{}
}

var invoiceCollection = database.OpenCollection(database.Client, "invoice")

func GetInvoices(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := invoiceCollection.Find(ctx, bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing invoice items"})
	}
	defer result.Close(ctx)

	var allInvoices []bson.M
	if err := result.All(ctx, &allInvoices); err != nil {
		log.Fatal(err)
	}
	return c.JSON(allInvoices)
}

func GetInvoice(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	invoiceId := c.Params("invoice_id")
	var invoice models.Invoice

	err := invoiceCollection.FindOne(ctx, bson.M{"invoice_id": invoiceId}).Decode(&invoice)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing invoice item"})
	}

	var invoiceView InvoiceViewFormat

	allOrderItems, err := ItemsByOrder(invoice.Order_id)
	invoiceView.Order_id = invoice.Order_id
	invoiceView.Payment_due_date = invoice.Payment_due_date

	invoiceView.Payment_method = "null"
	if invoice.Payment_method != nil {
		invoiceView.Payment_method = *invoice.Payment_method
	}

	invoiceView.Invoice_id = invoice.Invoice_id
	invoiceView.Payment_status = *&invoice.Payment_status
	invoiceView.Payment_due = allOrderItems[0]["payment_due"]
	invoiceView.Table_number = allOrderItems[0]["table_number"]
	invoiceView.Order_details = allOrderItems[0]["order_items"]

	return c.JSON(invoiceView)
}

func CreateInvoice(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice

	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var order models.Order

	err := orderCollection.FindOne(ctx, bson.M{"order_id": invoice.Order_id}).Decode(&order)
	if err != nil {
		msg := fmt.Sprintf("message: Order was not found")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}
	status := "PENDING"
	if invoice.Payment_status == nil {
		invoice.Payment_status = &status
	}

	invoice.Payment_due_date = time.Now().AddDate(0, 0, 1)
	invoice.Created_at = time.Now()
	invoice.Updated_at = time.Now()
	invoice.ID = primitive.NewObjectID()
	invoice.Invoice_id = invoice.ID.Hex()

	validationErr := validate.Struct(invoice)
	if validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	result, insertErr := invoiceCollection.InsertOne(ctx, invoice)
	if insertErr != nil {
		msg := fmt.Sprintf("invoice item was not created")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.JSON(result)
}

func UpdateInvoice(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var invoice models.Invoice
	invoiceId := c.Params("invoice_id")

	if err := c.BodyParser(&invoice); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	filter := bson.M{"invoice_id": invoiceId}

	var updateObj primitive.D

	if invoice.Payment_method != nil {
		updateObj = append(updateObj, bson.E{"payment_method", invoice.Payment_method})
	}

	if invoice.Payment_status != nil {
		updateObj = append(updateObj, bson.E{"payment_status", invoice.Payment_status})
	}

	invoice.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{"updated_at", invoice.Updated_at})

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	status := "PENDING"
	if invoice.Payment_status == nil {
		invoice.Payment_status = &status
	}

	result, err := invoiceCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)
	if err != nil {
		msg := fmt.Sprintf("invoice item update failed")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.JSON(result)
}
