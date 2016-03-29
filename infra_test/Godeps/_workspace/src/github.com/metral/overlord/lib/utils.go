package lib

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/metral/goutils"
)

// Get the IP address of the docker host as this is run from within container
func getDockerHostIP() string {
	cmd := fmt.Sprintf("netstat -nr | grep '^0\\.0\\.0\\.0' | awk '{print $2}'")
	out, err := exec.Command("sh", "-c", cmd).Output()
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

	ip := string(out)
	ip = strings.Replace(ip, "\n", "", -1)
	return ip
}

func removeOverlord(nodes *ResultNodes) {
	var fleetMachine FleetMachine
	n := *nodes

	for i, node := range n {
		WaitForMetadata(&node, &fleetMachine)
		if fleetMachine.Metadata["kubernetes_role"] == "overlord" {
			n = append(n[:i], n[i+1:]...)
			*nodes = n
			break
		}
	}

}

func IsMaster(fleetMachine *FleetMachine) bool {
	role := fleetMachine.Metadata["kubernetes_role"]

	switch role {
	case "master":
		return true
	}
	return false
}

func IsMinion(fleetMachine *FleetMachine) bool {
	role := fleetMachine.Metadata["kubernetes_role"]

	switch role {
	case "minion":
		return true
	}
	return false
}
