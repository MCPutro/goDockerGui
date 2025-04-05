package handler

import "github.com/gofiber/fiber/v2"

type ImageHandler interface {
	Show(c *fiber.Ctx) error
	Pull(c *fiber.Ctx) error
	Remove(c *fiber.Ctx) error
}
