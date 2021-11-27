package api

import (
	"quicksearch/indexer"
	"quicksearch/processor"

	"github.com/gofiber/fiber/v2"
)

func insertPost(c *fiber.Ctx) error {
	posts := new([]processor.Results)
	if err := c.BodyParser(posts); err != nil {
		c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
	}

	for _, post := range *posts {
		processor.QChan <- &post
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})
	return nil
}

func deletePost(c *fiber.Ctx) error {
	type inp struct {
		URLs []string `json:"urls"`
	}
	urls := new(inp)
	if err := c.BodyParser(urls); err != nil {
		c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	for _, url := range urls.URLs {
		if err := indexer.I.Store.DeletePost(&url); err != nil {
			c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
				"ok":    false,
				"error": err.Error(),
			})
			return nil
		}
	}
	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})

	return nil
}

func setSiteTokens(c *fiber.Ctx) error {
	type inp struct {
		Site   string             `json:"site"`
		Tokens map[string]float32 `json:"tokens"`
	}
	input := new(inp)
	if err := c.BodyParser(input); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	if err := indexer.I.Store.SetSiteTokens(&input.Site, &input.Tokens); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})

	return nil
}

func setTokensH(c *fiber.Ctx) error {
	type inp struct {
		URL    string             `json:"url"`
		Tokens map[string]float32 `json:"tokens"`
	}
	input := new(inp)
	if err := c.BodyParser(input); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	if err := indexer.I.Store.SetPostHTokens(&input.URL, &input.Tokens); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})

	return nil
}

func query(c *fiber.Ctx) error {
	type inp struct {
		Query        string  `json:"query"`
		Limit        uint32  `json:"limit"`
		Offset       uint32  `json:"offset"`
		Lang         *string `json:"lang"`
		ScoreDetails *bool   `json:"score_details"`
	}
	input := new(inp)

	if err := c.BodyParser(input); err != nil {
		c.Status(fiber.ErrBadRequest.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	if input.Lang == nil {
		input.Lang = new(string)
		*input.Lang = "en"
	}

	posts := new([]indexer.Post)
	if err := indexer.I.QueryFullText(&input.Query, input.Lang, input.Limit, input.Offset, posts); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
	}
	{
		// display score details
		if input.ScoreDetails == nil || !*input.ScoreDetails {
			for _, post := range *posts {
				post.Entities = nil
				post.ExternalLinks = nil
				post.InternalLinks = nil
				post.Tokens = nil
				post.TokensH = nil
				post.ExternalSiteScores = nil
			}
		}
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok":    true,
		"posts": *posts,
	})

	return nil
}

func reset(c *fiber.Ctx) error {
	if err := indexer.I.Store.Reset(); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})
	return nil
}

func sync(c *fiber.Ctx) error {
	if err := indexer.I.Store.Sync(); err != nil {
		c.Status(fiber.ErrInternalServerError.Code).JSON(fiber.Map{
			"ok":    false,
			"error": err.Error(),
		})
		return nil
	}

	// all done
	c.Status(fiber.StatusOK).JSON(fiber.Map{
		"ok": true,
	})
	return nil
}
