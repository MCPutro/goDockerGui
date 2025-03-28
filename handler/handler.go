package handler

import (
	"docker-ui/model"
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Handler interface {
	LoadData(w http.ResponseWriter, r *http.Request)
	Image(w http.ResponseWriter, r *http.Request)
	Container(w http.ResponseWriter, r *http.Request)
	StopContainer(w http.ResponseWriter, r *http.Request)
	StartContainer(w http.ResponseWriter, r *http.Request)
	DeleteContainer(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	Log(w http.ResponseWriter, r *http.Request)
	Inspect(w http.ResponseWriter, r *http.Request)
}

type HandleImpl struct {
	DockerClient *client.Client
	template     *template.Template
	modelDocker  *model.Docker
}

var (
	HandleImplOnce sync.Once
	Handle         Handler
	noWaitTimeout  = 0
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
	http.Redirect(w, r, "/image", http.StatusFound)
}

func (h *HandleImpl) Image(w http.ResponseWriter, r *http.Request) {
	h.modelDocker.Pid = os.Getpid()
	h.modelDocker.Images = nil
	h.modelDocker.Containers = nil

	//version, err := h.DockerClient.ServerVersion(r.Context())
	//if err != nil {
	//	fmt.Fprint(w, err)
	//	return
	//}
	//h.modelDocker.Version = version.Version

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
			Created:      time.Unix(img.Created, 0).Format("02 Jan 2006 15:04:05"),
			Size:         fmt.Sprintf("%.2f MB ", float64(img.Size)/(1024*1024)),
		}

		h.modelDocker.Images = append(h.modelDocker.Images, &dockerImage)
	}

	h.template.ExecuteTemplate(w, "image.html", h.modelDocker)
}

func (h *HandleImpl) Container(w http.ResponseWriter, r *http.Request) {
	h.modelDocker.Pid = os.Getpid()
	h.modelDocker.Images = nil
	h.modelDocker.Containers = nil

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
			//var tmp []string
			//for _, port := range ports {
			//	tmp = append(tmp, fmt.Sprintf("%s:%d -> %d", port.IP, port.PrivatePort, port.PublicPort))
			//}
			//portDetails = strings.Join(tmp, ", ")
			port := ports[0]
			portDetails = fmt.Sprintf("%d : %d", port.PublicPort, port.PrivatePort)
		} else {
			portDetails = ""
		}

		dockerContainer := model.DockerContainer{
			ContainerIDShow: c.ID[:12],
			ContainerID:     c.ID,
			Image:           c.Image,
			Status:          c.Status,
			Created:         time.Unix(c.Created, 0).Format("02 Jan 2006 15:04:05"),
			Port:            portDetails,
			Name:            strings.TrimPrefix(c.Names[0], "/"),
			State:           c.State,
		}

		h.modelDocker.Containers = append(h.modelDocker.Containers, &dockerContainer)
	}

	h.template.ExecuteTemplate(w, "container.html", h.modelDocker)
}

func (h *HandleImpl) StopContainer(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	err := h.DockerClient.ContainerStop(r.Context(), containerId, container.StopOptions{Timeout: &noWaitTimeout})

	response := make(map[string]interface{})
	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["success"] = false
		response["message"] = err.Error()
	} else {
		w.WriteHeader(http.StatusOK)
		response["success"] = true
		response["message"] = "success"
	}

	json.NewEncoder(w).Encode(response)
}

func (h *HandleImpl) StartContainer(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	err := h.DockerClient.ContainerStart(r.Context(), containerId, container.StartOptions{})

	response := make(map[string]interface{})
	w.Header().Add("Content-Type", "application/json")

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		response["success"] = false
		response["message"] = err.Error()
	} else {
		w.WriteHeader(http.StatusOK)
		response["success"] = true
		response["message"] = "success"
	}

	json.NewEncoder(w).Encode(response)
}

func (h *HandleImpl) DeleteContainer(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	if err := h.DockerClient.ContainerStop(r.Context(), containerId, container.StopOptions{Timeout: &noWaitTimeout}); err != nil {
		log.Printf("Unable to stop container %s: %s", containerId, err)
	}

	err := h.DockerClient.ContainerRemove(r.Context(), containerId, container.RemoveOptions{RemoveVolumes: true, Force: true})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}

}

func (h *HandleImpl) Ping(w http.ResponseWriter, r *http.Request) {
	h.template.ExecuteTemplate(w, "index.html", "Hello Template Caching")
}

func (h *HandleImpl) Log(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	logs, err := h.DockerClient.ContainerLogs(r.Context(), containerId, container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "100"})
	if err != nil {
		log.Fatal(err)
	}
	defer logs.Close()

	if all, err := io.ReadAll(logs); err == nil {
		//stringLogs := string(all)

		lines := strings.Split(string(all), "\n")

		var logOutput string
		for _, line := range lines {
			if len(line) > 0 {
				logOutput += line[8:] + "<br>"
			}
		}

		h.template.ExecuteTemplate(w, "log.html", template.HTML(logOutput))
	}

	////h.template.ExecuteTemplate(w, "log.html", template.HTML(data))
	//// Read the logs and convert to a string
	//var logOutput string
	//scanner := bufio.NewScanner(logs)
	//for scanner.Scan() {
	//	logOutput += scanner.Text()[8:] + "\n"
	//}
	//if err := scanner.Err(); err != nil {
	//	log.Fatalf("Error reading logs: %v", err)
	//}
	////cleaned := regexp.MustCompile(`[^[:print:]\n\r\t]`).ReplaceAllString(logOutput, "")
	//fmt.Println(logOutput)
}

func (h *HandleImpl) Inspect(w http.ResponseWriter, r *http.Request) {
	containerId := r.PathValue("containerId")
	if containerId == "" {
		http.Error(w, "Missing tech parameter", http.StatusBadRequest)
		return
	}

	// Inspect the container
	containerInspect, err := h.DockerClient.ContainerInspect(r.Context(), containerId)
	if err != nil {
		log.Fatal(err)
	}

	m := make(map[string]map[string]interface{})

	if m["Environment"] == nil {
		m["Environment"] = make(map[string]interface{})
	}
	for _, e := range containerInspect.Config.Env {
		split := strings.Split(e, "=")
		m["Environment"][split[0]] = split[1]
	}

	if m["Mounts"] == nil {
		m["Mounts"] = make(map[string]interface{})
	}
	for _, mount := range containerInspect.Mounts {
		//fmt.Println(mount.Type, " -> ", mount.Source, " -> ", mount.Destination)
		m["Mounts"][mount.Destination] = mount.Source
	}

	//fmt.Println("Port : ")
	if m["Ports"] == nil {
		m["Ports"] = make(map[string]interface{})
	}
	for port, bindings := range containerInspect.HostConfig.PortBindings {
		m["Ports"][port.Port()+"/"+port.Proto()] = cek(bindings)
	}

	//fmt.Println(m)

	h.template.ExecuteTemplate(w, "inspect.html", m)

}

func cek(binding []nat.PortBinding) string {
	var temp []string
	for _, portBinding := range binding {
		if portBinding.HostIP != "" {
			temp = append(temp, fmt.Sprintf("%s:%s", portBinding.HostIP, portBinding.HostPort))
		} else {
			temp = append(temp, fmt.Sprintf("localhost:%s", portBinding.HostPort))
		}
	}
	return strings.Join(temp, ", ")
}
