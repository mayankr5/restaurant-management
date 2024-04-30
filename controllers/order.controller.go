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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

var orderCollection *mongo.Collection = database.OpenCollection(database.Client, "order")

func GetOrders(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := orderCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing order items"})
	}
	defer result.Close(ctx)

	var allOrders []bson.M
	if err = result.All(ctx, &allOrders); err != nil {
		log.Fatal(err)
	}
	return c.JSON(allOrders)
}

func GetOrder(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	orderId := c.Params("order_id")
	var order models.Order

	err := orderCollection.FindOne(ctx, bson.M{"order_id": orderId}).Decode(&order)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the orders"})
	}
	return c.JSON(order)
}

func CreateOrder(c *fiber.Ctx) error {
	var table models.Table
	var order models.Order

	if err := c.BodyParser(&order); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	validationErr := validate.Struct(order)

	if validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	if order.Table_id != nil {
		err := tableCollection.FindOne(ctx, bson.M{"table_id": order.Table_id}).Decode(&table)
		if err != nil {
			msg := fmt.Sprintf("message:Table was not found")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}
	}

	order.Created_at = time.Now()
	order.Updated_at = time.Now()
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	result, insertErr := orderCollection.InsertOne(ctx, order)

	if insertErr != nil {
		msg := fmt.Sprintf("order item was not created")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.JSON(result)
}

func UpdateOrder(c *fiber.Ctx) error {
	var table models.Table
	var order models.Order

	var updateObj primitive.D

	orderId := c.Params("order_id")
	if err := c.BodyParser(&order); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if order.Table_id != nil {
		err := menuCollection.FindOne(ctx, bson.M{"tabled_id": order.Table_id}).Decode(&table)
		if err != nil {
			msg := fmt.Sprintf("message:Menu was not found")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}
		updateObj = append(updateObj, bson.E{"menu", order.Table_id})
	}

	order.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{"updated_at", order.Updated_at})

	upsert := true

	filter := bson.M{"order_id": orderId}
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := orderCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$st", updateObj},
		},
		&opt,
	)

	if err != nil {
		msg := fmt.Sprintf("order item update failed")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.JSON(result)
}

func OrderItemOrderCreator(order models.Order) string {

	order.Created_at = time.Now()
	order.Updated_at = time.Now()
	order.ID = primitive.NewObjectID()
	order.Order_id = order.ID.Hex()

	orderCollection.InsertOne(ctx, order)

	return order.Order_id
}
