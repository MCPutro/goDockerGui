package model

type Docker struct {
	Pid        int
	Version    string
	Images     []*DockerImage
	Containers []*DockerContainer
}

type DockerImage struct {
	RepositoryID string
	Tag          string
	ImageID      string
	Created      string
	Size         string
}

type DockerContainer struct {
	ContainerID string
	Image       string
	Status      string
	Created     string
	Port        string
	Name        string
	State       string
}
