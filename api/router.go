package api

import "github.com/gofiber/fiber/v2"

func Routes(router *fiber.Router) {
	(*router).Post("/data/insert", insertPost)
	(*router).Post("/data/delete", deletePost)
	(*router).Delete("/data/reset", reset)
	(*router).Post("/q", query)
	(*router).Get("/data/sync", sync)
}
