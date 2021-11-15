package api

import "github.com/gofiber/fiber/v2"

func Routes(router *fiber.Router) {
	(*router).Post("/data/insert", insertPost)
	(*router).Delete("/data/reset", reset)
}
