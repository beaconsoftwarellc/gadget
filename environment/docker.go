package environment

import (
	"os/exec"
	"strconv"
	"strings"

	"github.com/samalba/dockerclient"
)

const dockerSocket = "docker://"

// IsDockerSocket checks that the string has the correct prefix for a docker service lookup
func IsDockerSocket(value string) bool {
	return strings.HasPrefix(value, dockerSocket)
}

// LookupContainerID checks the /etc/hostname for the Container ID
// The default docker behavior is to place the Short Container ID in /etc/hostname
func LookupContainerID() string {
	cmd := exec.Command("cat", "/etc/hostname")
	out, _ := cmd.Output()
	containerID := strings.TrimSpace(string(out))
	return containerID
}

// LookupHostPort uses the docker socket to query the Host Port
// The `docker info` command can provide the host port information, the dockerclient is a wrapper for docker calls.
// The docker container needs to be configured with the dockerSocket mounted
func LookupHostPort(localPort int, socket string, hostPort int) int {
	if !IsDockerSocket(socket) {
		return hostPort
	}
	docker, _ := dockerclient.NewDockerClient(strings.Replace(socket, dockerSocket, "unix://", 1), nil)
	containerID := LookupContainerID()
	info, _ := docker.InspectContainer(containerID)
	port := info.NetworkSettings.Ports[strconv.Itoa(localPort)+"/tcp"][0].HostPort
	portNumber, _ := strconv.Atoi(port)
	return portNumber
}
