package handler

import (
	"docker-ui/model"
	"docker-ui/utils/constants"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/gofiber/fiber/v2"
	"io"
	"os"
	"strings"
	"time"
)

type imageHandlerImpl struct {
	dockerClient *client.Client
}

func NewImageHandler(dockerClient *client.Client) ImageHandler {
	return &imageHandlerImpl{dockerClient: dockerClient}
}

var (
	modelImage []*model.DockerImage
)

func (i *imageHandlerImpl) Show(c *fiber.Ctx) error {
	modelImage = nil

	imageList, err := i.dockerClient.ImageList(c.Context(), image.ListOptions{
		All: true,
	})

	for _, img := range imageList {
		split := strings.Split(img.RepoTags[0], ":")

		dockerImage := model.DockerImage{
			RepositoryID: split[0],
			Tag:          split[1],
			ImageID:      strings.TrimPrefix(img.ID, "sha256:")[:12],
			Created:      time.Unix(img.Created, 0).Format("02 Jan 2006 15:04:05"),
			Size:         fmt.Sprintf("%.2f MB ", float64(img.Size)/(1024*1024)),
		}

		modelImage = append(modelImage, &dockerImage)
	}

	return c.Render("template/image", fiber.Map{
		constants.ERRORS: err,
		constants.PID:    os.Getpid(),
		constants.IMAGES: modelImage,
	})
}

func (i *imageHandlerImpl) Pull(c *fiber.Ctx) error {
	var body map[string]string

	err := c.BodyParser(&body)
	if err != nil {
	}

	s := body["imageId"]

	if s == "" {
		return errors.New("cant't pull image")
	}

	if strings.HasPrefix(s, "docker pull") {
		s = strings.TrimPrefix(s, "docker pull")
	}

	reader, err := i.dockerClient.ImagePull(c.Context(), strings.TrimSpace(s), image.PullOptions{})
	defer reader.Close()

	io.Copy(os.Stdout, reader)

	//readAll, err := io.ReadAll(pull)
	//lines := strings.Split(string(readAll), "\n")

	//var logOutput string
	//for _, line := range readAll {
	//	if len(line) > 0 {
	//		logOutput += line[8:] + "<br>"
	//	}
	//}

	return c.SendString("ok")
}

func (i *imageHandlerImpl) Remove(c *fiber.Ctx) error {
	return c.SendString("remove belum jadi")
}
