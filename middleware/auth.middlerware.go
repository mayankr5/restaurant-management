package middleware

import (
	"fmt"
	"net/http"

	"github.com/gofiber/fiber/v2"
	helper "github.com/mayankr5/v1/restaurant-management/helpers"
)

func Authentication() fiber.Handler {
	return func(c *fiber.Ctx) error {
		clientToken := c.Get("token")
		if clientToken == "" {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("No Authorization header provided")})
		}

		claims, err := helper.ValidateToken(clientToken)
		if err != "" {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{"error": err})
		}

		c.Locals("email", claims.Email)
		c.Locals("first_name", claims.First_name)
		c.Locals("last_name", claims.Last_name)
		c.Locals("uid", claims.Uid)

		return c.Next()
	}
}
