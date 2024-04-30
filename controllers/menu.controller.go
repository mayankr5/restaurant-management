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

var menuCollection *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	result, err := menuCollection.Find(context.TODO(), bson.M{})
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while listing the menu items"})
	}
	defer result.Close(ctx)

	var allMenus []bson.M
	if err := result.All(ctx, &allMenus); err != nil {
		log.Fatal(err)
	}
	return c.JSON(allMenus)
}

func GetMenu(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	menuId := c.Params("menu_id")
	var menu models.Menu

	err := menuCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": "error occurred while fetching the menu"})
	}
	return c.JSON(menu)
}

func CreateMenu(c *fiber.Ctx) error {
	var menu models.Menu
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	if err := c.BodyParser(&menu); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	validationErr := validate.Struct(menu)
	if validationErr != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": validationErr.Error()})
	}

	menu.Created_at = time.Now()
	menu.Updated_at = time.Now()
	menu.ID = primitive.NewObjectID()
	menu.Menu_id = menu.ID.Hex()

	result, insertErr := menuCollection.InsertOne(ctx, menu)
	if insertErr != nil {
		msg := fmt.Sprintf("Menu item was not created")
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
	}
	return c.JSON(result)
}

func UpdateMenu(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var menu models.Menu

	if err := c.BodyParser(&menu); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	menuId := c.Params("menu_id")
	filter := bson.M{"menu_id": menuId}

	var updateObj primitive.D

	if menu.Start_Date != nil && menu.End_Date != nil {
		if !inTimeSpan(*menu.Start_Date, *menu.End_Date, time.Now()) {
			msg := "kindly retype the time"
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}

		updateObj = append(updateObj, bson.E{"start_date", menu.Start_Date})
		updateObj = append(updateObj, bson.E{"end_date", menu.End_Date})

		if menu.Name != "" {
			updateObj = append(updateObj, bson.E{"name", menu.Name})
		}
		if menu.Category != "" {
			updateObj = append(updateObj, bson.E{"name", menu.Category})
		}

		menu.Updated_at = time.Now()
		updateObj = append(updateObj, bson.E{"updated_at", menu.Updated_at})

		upsert := true

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := menuCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{"$set", updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := "Menu update failed"
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": msg})
		}

		return c.JSON(result)
	}
	return nil
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(start)
}
