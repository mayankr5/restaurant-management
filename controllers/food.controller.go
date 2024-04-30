package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/mayankr5/v1/restaurant-management/database"
	"github.com/mayankr5/v1/restaurant-management/models"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

func GetFoods(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
	if err != nil || recordPerPage < 1 {
		recordPerPage = 10
	}

	page, err := strconv.Atoi(c.Query("page"))
	if err != nil || page < 1 {
		page = 1
	}

	startIndex := (page - 1) * recordPerPage
	startIndex, err = strconv.Atoi(c.Query("startIndex"))

	matchStage := bson.D{{"$match", bson.D{{}}}}
	groupStage := bson.D{{"$group", bson.D{{"_id", bson.D{{"_id", "null"}}}, {"total_count", bson.D{{"$sum", 1}}}, {"data", bson.D{{"$push", "$$ROOT"}}}}}}
	projectStage := bson.D{
		{
			"$project", bson.D{
				{"_id", 0},
				{"total_count", 1},
				{"food_items", bson.D{{"$slice", []interface{}{"$data", startIndex, recordPerPage}}}},
			}}}

	result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{
		matchStage, groupStage, projectStage})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing food items"})
	}
	defer result.Close(ctx)

	var allFoods []bson.M
	if err := result.All(ctx, &allFoods); err != nil {
		log.Fatal(err)
	}
	return c.JSON(allFoods[0])
}

func GetFood(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	foodId := c.Params("food_id")
	var food models.Food

	err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the food item"})
	}
	return c.JSON(food)
}

func CreateFood(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	var food models.Food

	if err := c.BodyParser(&food); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	validationErr := validate.Struct(food)
	if validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
	if err != nil {
		msg := fmt.Sprintf("menu was not found")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}

	food.Created_at = time.Now()
	food.Updated_at = time.Now()
	food.ID = primitive.NewObjectID()
	food.Food_id = food.ID.Hex()
	price := toFixed(*food.Price, 2)
	food.Price = &price

	result, insertErr := foodCollection.InsertOne(ctx, food)
	if insertErr != nil {
		msg := fmt.Sprintf("Food item was not created")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}
	return c.JSON(result)
}

func UpdateFood(c *fiber.Ctx) error {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu
	var food models.Food

	foodId := c.Params("food_id")

	if err := c.BodyParser(&food); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	var updateObj primitive.D

	if food.Name != nil {
		updateObj = append(updateObj, bson.E{"name", food.Name})
	}

	if food.Price != nil {
		updateObj = append(updateObj, bson.E{"price", food.Price})
	}

	if food.Food_image != nil {
		updateObj = append(updateObj, bson.E{"food_image", food.Food_image})
	}

	if food.Menu_id != nil {
		err := menuCollection.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		if err != nil {
			msg := fmt.Sprintf("message:Menu was not found")
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}
		updateObj = append(updateObj, bson.E{"menu", food.Price})
	}

	food.Updated_at = time.Now()
	updateObj = append(updateObj, bson.E{"updated_at", food.Updated_at})

	upsert := true
	filter := bson.M{"food_id": foodId}

	opt := options.UpdateOptions{
		Upsert: &upsert,
	}

	result, err := foodCollection.UpdateOne(
		ctx,
		filter,
		bson.D{
			{"$set", updateObj},
		},
		&opt,
	)

	if err != nil {
		msg := fmt.Sprint("foot item update failed")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}
	return c.JSON(result)
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	res := float64(round(num*output)) / output
	return res
}
