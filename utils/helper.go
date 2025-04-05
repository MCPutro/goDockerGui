package utils

import (
	"fmt"
	"github.com/docker/go-connections/nat"
	"strings"
)

func JonPorts(binding []nat.PortBinding) string {
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
