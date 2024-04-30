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

var tableCollection *mongo.Collection = database.OpenCollection(database.Client, "table")

func GetTables(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := orderCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing table items"})
	}
	defer result.Close(ctx)

	var allTables []bson.M
	if err = result.All(ctx, &allTables); err != nil {
		log.Fatal(err)
	}
	return c.JSON(allTables)
}

func GetTable(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	tableId := c.Params("table_id")
	var table models.Table

	err := tableCollection.FindOne(ctx, bson.M{"table_id": tableId}).Decode(&table)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the tables"})
	}
	return c.JSON(table)
}

func CreateTable(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table

	if err := c.BodyParser(&table); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	validationErr := validate.Struct(table)
	if validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	table.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	table.ID = primitive.NewObjectID()
	table.Table_id = table.ID.Hex()

	result, insertErr := tableCollection.InsertOne(ctx, table)
	if insertErr != nil {
		msg := fmt.Sprintf("Table item was not created")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.JSON(result)
}

func UpdateTable(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var table models.Table

	tableId := c.Params("table_id")

	if err := c.BodyParser(&table); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var updateObj primitive.D

	if table.Number_of_guests != nil {
		updateObj = append(updateObj, bson.E{"number_of_guests", table.Number_of_guests})
	}

	if table.Table_number != nil {
		updateObj = append(updateObj, bson.E{"table_number", table.Table_number})
	}

	table.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	upsert := true
	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	filter := bson.M{"table_id": tableId}

	result, err := tableCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	if err != nil {
		msg := fmt.Sprintf("table item update failed")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	return c.JSON(result)
}
