package main

import (
	"docker-ui/handler"
	"embed"

	"github.com/docker/docker/client"

	//"github.com/gofiber/fiber/v2"
	//"github.com/gofiber/fiber/v2/middleware/filesystem"
	//goHtml "github.com/gofiber/template/html/v2"
	"html/template"
	"log"
	"net/http"
)

//go:embed template/*.html
//go:embed template/fragment/*
var templates2 embed.FS

func main() {

	var myTemplates = template.Must(template.ParseFS(templates2, "template/*.html", "template/fragment/*.html"))
	// Membuat koneksi ke Docker daemon
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	handleImpl := handler.NewHandleImpl(cli, myTemplates)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", handleImpl.LoadData)
	mux.HandleFunc("GET /container/stop/{containerId}", handleImpl.StopContainer)
	mux.HandleFunc("GET /container/start/{containerId}", handleImpl.StartContainer)
	mux.HandleFunc("DELETE /container/delete/{containerId}", handleImpl.DeleteContainer)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("template/static"))))

	mux.HandleFunc("GET /image", handleImpl.Image)
	mux.HandleFunc("GET /container", handleImpl.Container)
	mux.HandleFunc("GET /container/log/{containerId}", handleImpl.Log)
	mux.HandleFunc("GET /container/inspect/{containerId}", handleImpl.Inspect)
	mux.HandleFunc("POST /container/collection/delete", handleImpl.DeleteContainerCollection)

	// Start the server
	port := ":5000"
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	log.Println("Listening... http://localhost" + port)
	server.ListenAndServe() // Run the http server
}

//func main2() {
//	engine := goHtml.NewFileSystem(http.FS(templates2), ".html")
//
//	app := fiber.New(fiber.Config{Views: engine})
//	//app := fiber.New()
//	app.Use("/static/", filesystem.New(filesystem.Config{
//		Root: http.Dir("./template/static"),
//	}))
//
//	app.Get("/fiber1", func(c *fiber.Ctx) error {
//		return c.Render("template/container", fiber.Map{})
//	})
//
//	err := app.Listen(":3000")
//	if err != nil {
//		panic(err)
//	}
//}
