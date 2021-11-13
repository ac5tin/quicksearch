package api

import (
	"quicksearch/indexer"

	"github.com/gofiber/fiber/v2"
)

func insertPost(c *fiber.Ctx) error {
	posts := new([]indexer.Post)
	if err := c.BodyParser(posts); err != nil {
		c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})

		for _, p := range *posts {
			if err := indexer.I.Store.InsertPost(&p); err != nil {
				c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
					"ok":    false,
					"error": err.Error(),
				})
				return nil
			}
		}
		return nil
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})
	return nil
}

func deletePost(c *fiber.Ctx) error {
	return nil
}

func query(c *fiber.Ctx) error {
	return nil
}
