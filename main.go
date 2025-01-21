package main

import (
	"docker-ui/handler"
	"embed"
	"github.com/docker/docker/client"
	"html/template"
	"log"
	"net/http"
)

//go:embed template/*.html
var templates2 embed.FS

func main() {

	var myTemplates = template.Must(template.ParseFS(templates2, "template/*.html"))
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

	// Start the server
	port := ":9999"
	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}
	log.Println("Listening...")
	server.ListenAndServe() // Run the http server
}
