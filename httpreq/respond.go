package httpreq

import "github.com/gofiber/fiber/v2"

type Paging struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Total int `json:"total"`
}

func JSONOK(c *fiber.Ctx, data any) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"data": data})
}

func JSONList(c *fiber.Ctx, data any, p Paging) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"data":   data,
		"paging": p,
	})
}

func Err(c *fiber.Ctx, status int, msg string) error {
	return c.Status(status).JSON(fiber.Map{
		"error": fiber.Map{"message": msg},
	})
}

func ErrInternal(c *fiber.Ctx, err error) error {
	return Err(c, fiber.StatusInternalServerError, err.Error())
}
