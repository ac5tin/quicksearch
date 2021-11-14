package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"quicksearch/api"
	"quicksearch/db"
	"quicksearch/indexer"

	gr "github.com/ac5tin/goredis"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// init indexer
	// -- init store for indexer
	// ---- initialise postgres
	log.Println("... initialising database connection") // debug
	pg, err := db.Db(os.Getenv("DB_STRING"), os.Getenv("DB_SCHEMA"))
	if err != nil {
		panic(err)
	}
	log.Println("> Successfully established database connection <") // debug
	// ---- initialise redis
	log.Println("... initialising redis connection") // debug
	rc := gr.NewRedisClient(fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")), 0, "")
	log.Println("> Successfully established redis connection <") // debug

	// ---- initialise store
	s := indexer.NewStore(rc, pg)
	indexer.I = new(indexer.Indexer)
	indexer.I.Store = &s

	// start REST  API server
	port := flag.Uint("p", 7898, "Port number")
	prefork := flag.Bool("prefork", false, "Prefork")

	flag.Parse()

	app := fiber.New(fiber.Config{
		Prefork: *prefork,
	})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(compress.New())
	app.Use(cors.New())

	// ===== API ROUTES =====
	app.Get("/ping", func(c *fiber.Ctx) error { c.Status(fiber.StatusOK).Send([]byte("pong")); return nil })
	apiGroup := app.Group("/api")
	api.Routes(&apiGroup)

	log.Println(fmt.Sprintf("Listening on PORT %d", *port))
	if err := app.Listen(fmt.Sprintf(":%d", *port)); err != nil {
		log.Fatal(err.Error())
	}
}
