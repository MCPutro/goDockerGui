package handler

import "github.com/gofiber/fiber/v2"

type ContainerHandler interface {
	Show(c *fiber.Ctx) error

	Action(c *fiber.Ctx) error

	Inspect(c *fiber.Ctx) error
	Log(c *fiber.Ctx) error

	BatchDelete(c *fiber.Ctx) error
}
