package main

import (
	"docker-ui/handler"
	"embed"

	"github.com/docker/docker/client"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	goHtml "github.com/gofiber/template/html/v2"
	"html/template"
	"log"
	"net/http"
)

//go:embed template/*
//go:embed template/fragment/*
var templates2 embed.FS

func main() {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
		return
	}
	dockerHandler := handler.NewDockerHandler(cli)
	imageHandler := handler.NewImageHandler(cli)

	engine := goHtml.NewFileSystem(http.FS(templates2), ".html")
	engine.AddFunc(
		"unescape", func(s string) template.HTML {
			return template.HTML(s)
		},
	)

	app := fiber.New(fiber.Config{Views: engine})
	//app := fiber.New()

	app.Use("/static/", filesystem.New(filesystem.Config{
		Root: http.Dir("./template/static"),
	}))

	app.Get("/container", dockerHandler.Show)
	app.Put("/container/:action/:containerId", dockerHandler.Action)
	app.Get("/container/inspect/:containerId", dockerHandler.Inspect)
	app.Get("/container/log/:containerId", dockerHandler.Log)
	app.Post("/container/batch-delete", dockerHandler.BatchDelete)

	app.Get("/image", imageHandler.Show)
	app.Post("/image", imageHandler.Pull)

	app.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/image") })

	err = app.Listen(":5000")
	if err != nil {
		panic(err)
	}
}
