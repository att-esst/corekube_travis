package lib

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

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
		waitForMetadata(&node, &fleetMachine)
		if fleetMachine.Metadata["kubernetes_role"] == "overlord" {
			n = append(n[:i], n[i+1:]...)
			*nodes = n
			break
		}
	}

}

func Main() {
	fleetResult := Result{}
	var f *Result = &fleetResult

	setMachinesSeen([]string{})

	// Get Fleet machines
	var masterFleetMachine FleetMachine
	var createdFiles, allMachinesSeen []string
	for {
		getFleetMachines(f)

		totalSeen := len(allMachinesSeen)
		log.Printf("------------------------------------------------------------")
		log.Printf("Current # of machines seen/deployed to: (%d)\n", totalSeen)

		totalMachines := len(f.Node.Nodes)
		log.Printf("------------------------------------------------------------")
		log.Printf("Current # of machines discovered: (%d)\n", totalMachines)

		var fleetMachine FleetMachine
		// Get Fleet machines metadata
		for _, resultNode := range f.Node.Nodes {
			waitForMetadata(&resultNode, &fleetMachine)

			switch fleetMachine.Metadata["kubernetes_role"] {
			case "master":
				masterFleetMachine = fleetMachine
			}

			if !machineSeen(allMachinesSeen, fleetMachine.ID) &&
				masterFleetMachine.ID != "" {

				log.Printf("------------------------------------------------------------")
				log.Printf("Found machine:\n")
				fleetMachine.PrintString()

				if isMaster(&fleetMachine) || isMinion(&fleetMachine) {
					allMachinesSeen = append(allMachinesSeen, fleetMachine.ID)

					if isMaster(&fleetMachine) {
						createdFiles = createMasterUnits(&fleetMachine)
					} else if isMinion(&fleetMachine) {
						createdFiles = createMinionUnits(
							&masterFleetMachine, &fleetMachine)
					}
					for _, file := range createdFiles {
						if !unitFileCompleted(file) {
							startUnitFile(file)
							waitUnitFileComplete(file)
						}
					}

					//if isMinion(&fleetMachine) {
					//	registerKNodes(&masterFleetMachine, &fleetMachine)
					//}

					setMachinesSeen(allMachinesSeen)
				}
			}
		}

		time.Sleep(1 * time.Second)
		allMachinesSeen = getMachinesSeen()
	}
}
