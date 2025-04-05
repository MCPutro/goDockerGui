package handler

import (
	"docker-ui/model"
	"docker-ui/utils"
	"docker-ui/utils/constants"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

type containerHandlerImpl struct {
	dockerClient *client.Client
}

var (
	dockerHandlerOnce sync.Once
	dockerHandler     ContainerHandler

	modelContainers []*model.DockerContainer

	timeout = constants.NoWaitTimeout
)

func NewDockerHandler(dockerClient *client.Client) ContainerHandler {
	dockerHandlerOnce.Do(func() {
		dockerHandler = &containerHandlerImpl{dockerClient: dockerClient}
	})
	return dockerHandler
}

func (d *containerHandlerImpl) Show(c *fiber.Ctx) error {
	modelContainers = nil

	containers, err := d.dockerClient.ContainerList(c.Context(), container.ListOptions{All: true})

	for _, co := range containers {
		ports := co.Ports
		var portDetails string
		if len(ports) > 0 {
			portDetails = fmt.Sprintf("%d : %d", ports[0].PublicPort, ports[0].PrivatePort)
		} else {
			portDetails = ""
		}

		dockerContainer := model.DockerContainer{
			ContainerIDShow: co.ID[:12],
			ContainerID:     co.ID,
			Image:           co.Image,
			Status:          co.Status,
			Created:         time.Unix(co.Created, 0).Format("02 Jan 2006 15:04:05"),
			Port:            portDetails,
			Name:            strings.TrimPrefix(co.Names[0], "/"),
			State:           co.State,
		}

		modelContainers = append(modelContainers, &dockerContainer)
	}

	return c.Render("template/container", fiber.Map{
		constants.ERRORS:     err,
		constants.PID:        os.Getpid(),
		constants.CONTAINERS: modelContainers,
	})
}

func (d *containerHandlerImpl) Action(c *fiber.Ctx) error {
	action := c.Params("action", "")
	containerId := c.Params("containerId", "")

	if containerId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing container id")
	}

	var err error
	switch action {
	case "stop":
		err = d.dockerClient.ContainerStop(c.Context(), containerId, container.StopOptions{Timeout: &timeout})
	case "start":
		err = d.dockerClient.ContainerStart(c.Context(), containerId, container.StartOptions{})
	case "delete":
		err = d.dockerClient.ContainerRemove(c.Context(), containerId, container.RemoveOptions{RemoveVolumes: true, Force: true})
	default:
		err = errors.New("unknown action")
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.SendStatus(fiber.StatusOK)
}

func (d *containerHandlerImpl) Inspect(c *fiber.Ctx) error {
	containerId := c.Params("containerId", "")

	if containerId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing container id")
	}

	// Inspect the container
	containerInspect, err := d.dockerClient.ContainerInspect(c.Context(), containerId)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	m := make(map[string]map[string]interface{})

	if m[constants.ENV] == nil {
		m[constants.ENV] = make(map[string]interface{})
	}
	for _, e := range containerInspect.Config.Env {
		split := strings.Split(e, "=")
		m[constants.ENV][split[0]] = split[1]
	}

	if m[constants.MOUNTS] == nil {
		m[constants.MOUNTS] = make(map[string]interface{})
	}
	for _, mount := range containerInspect.Mounts {
		//fmt.Println(mount.Type, " -> ", mount.Source, " -> ", mount.Destination)
		m[constants.MOUNTS][mount.Destination] = mount.Source
	}

	//fmt.Println("Port : ")
	if m[constants.PORTS] == nil {
		m[constants.PORTS] = make(map[string]interface{})
	}
	for port, bindings := range containerInspect.HostConfig.PortBindings {
		m[constants.PORTS][port.Port()+"/"+port.Proto()] = utils.JonPorts(bindings)
	}

	return c.Render("template/inspect", fiber.Map{
		constants.ENV:    m[constants.ENV],
		constants.MOUNTS: m[constants.MOUNTS],
		constants.PORTS:  m[constants.PORTS],
	})
}

func (d *containerHandlerImpl) Log(c *fiber.Ctx) error {
	containerId := c.Params("containerId", "")

	if containerId == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Missing container id")
	}

	logs, err := d.dockerClient.ContainerLogs(c.Context(), containerId, container.LogsOptions{ShowStdout: true, ShowStderr: true, Tail: "100"})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	defer logs.Close()

	readAll, err := io.ReadAll(logs)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	lines := strings.Split(string(readAll), "\n")

	var logOutput string
	for _, line := range lines {
		if len(line) > 0 {
			logOutput += line[8:] + "<br>"
		}
	}

	return c.Render("template/log", fiber.Map{
		"data": logOutput,
	})
}

func (d *containerHandlerImpl) BatchDelete(c *fiber.Ctx) error {
	var err error
	var body map[string][]string
	err = c.BodyParser(&body)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString(err.Error())
	}

	for _, containerId := range body["containerIds"] {
		err = d.dockerClient.ContainerRemove(c.Context(), containerId, container.RemoveOptions{Force: true})

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
		}
	}

	return c.SendStatus(fiber.StatusOK)
}
