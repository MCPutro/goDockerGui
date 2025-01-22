package handler

import (
	"docker-ui/model"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Handler interface {
	LoadData(w http.ResponseWriter, r *http.Request)
	Image(w http.ResponseWriter, r *http.Request)
	StopContainer(w http.ResponseWriter, r *http.Request)
	StartContainer(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
}

type HandleImpl struct {
	DockerClient *client.Client
	template     *template.Template
	modelDocker  *model.Docker
}

var (
	HandleImplOnce sync.Once
	Handle         Handler
)

func NewHandleImpl(dockerClient *client.Client, template *template.Template) Handler {
	HandleImplOnce.Do(func() {
		Handle = &HandleImpl{
			DockerClient: dockerClient,
			template:     template,
			modelDocker:  &model.Docker{},
		}
	})
	return Handle
}

func (h *HandleImpl) LoadData(w http.ResponseWriter, r *http.Request) {
	h.modelDocker.Images = nil
	h.modelDocker.Containers = nil

	version, err := h.DockerClient.ServerVersion(r.Context())
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	h.modelDocker.Version = version.Version

	// imge
	imageList, err := h.DockerClient.ImageList(r.Context(), image.ListOptions{
		All: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	for _, img := range imageList {
		split := strings.Split(img.RepoTags[0], ":")

		dockerImage := model.DockerImage{
			RepositoryID: split[0],
			Tag:          split[1],
			ImageID:      strings.TrimPrefix(img.ID, "sha256:")[:12],
			Created:      time.Unix(img.Created, 0).Format("02/01/2006 15:04:05"),
			Size:         fmt.Sprintf("%.2f MB ", float64(img.Size)/(1024*1024)),
		}

		h.modelDocker.Images = append(h.modelDocker.Images, &dockerImage)
	}

	//container
	containers, err := h.DockerClient.ContainerList(r.Context(), container.ListOptions{All: true})
	if err != nil {
		log.Fatal(err)
	}

	for _, c := range containers {
		ports := c.Ports
		var portDetails string
		i := len(ports)
		if i > 0 {
			var tmp []string
			for _, port := range ports {
				tmp = append(tmp, fmt.Sprintf("%s:%d -> %d", port.IP, port.PrivatePort, port.PublicPort))
			}
			portDetails = strings.Join(tmp, ", ")
		} else {
			portDetails = "No ports mapped"
		}

		dockerContainer := model.DockerContainer{
			ContainerID: c.ID[:12],
			Image:       c.Image,
			Status:      c.Status,
			Created:     time.Unix(c.Created, 0).Format("02/01/2006 15:04:05"),
			Port:        portDetails,
			Name:        strings.TrimPrefix(c.Names[0], "/"),
			State:       c.State,
		}

		h.modelDocker.Containers = append(h.modelDocker.Containers, &dockerContainer)
	}

	h.template.ExecuteTemplate(w, "index.html", h.modelDocker)
}

func (h *HandleImpl) Image(w http.ResponseWriter, r *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (h *HandleImpl) StopContainer(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	err := h.DockerClient.ContainerStop(r.Context(), containerId, container.StopOptions{})
	if err != nil {
		//fmt.Fprintf(w, "Error stopping container: %v", err)
		h.template.ExecuteTemplate(w, "notif.html", fmt.Sprintf("Error stopping container: %v", err))
	} else {
		//fmt.Fprintf(w, "Container %s stopped successfully!", containerId)
		h.template.ExecuteTemplate(w, "notif.html", fmt.Sprintf("Container %s stopped successfully!", containerId))
	}
	//http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func (h *HandleImpl) StartContainer(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	err := h.DockerClient.ContainerStart(r.Context(), containerId, container.StartOptions{})
	if err != nil {
		//fmt.Fprintf(w, "Error starting container: %v", err)
		h.template.ExecuteTemplate(w, "notif.html", fmt.Sprintf("Error starting container: %v", err))
	} else {
		//fmt.Fprintf(w, "Container %s started successfully!", containerId)
		h.template.ExecuteTemplate(w, "notif.html", fmt.Sprintf("Container %s started successfully!", containerId))
	}
	//http.Redirect(w, r, "/", http.StatusPermanentRedirect)
}

func (h *HandleImpl) Ping(w http.ResponseWriter, r *http.Request) {
	h.template.ExecuteTemplate(w, "index.html", "Hello Template Caching")
}
