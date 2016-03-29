package lib

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/coreos/fleet/unit"
	"github.com/metral/goutils"
)

// Types Result, ResultNode, NodeResult & Node adapted from:
// https://github.com/coreos/fleet/blob/master/etcd/result.go
type Map map[string]interface{}

type FleetMachines []FleetMachine
type FleetMachine struct {
	ID             string
	PublicIP       string
	Metadata       Map
	Version        string
	TotalResources Map
}

type FleetUnitState struct {
	Hash               string
	MachineID          string
	Name               string
	SystemdActiveState string
	SystemdLoadState   string
	SystemdSubState    string
}

type FleetUnitStates struct {
	States []FleetUnitState
}

func (f FleetMachine) PrintString() {
	log.Printf("-- ID: %s\n", f.ID)
	log.Printf("-- IP: %s\n", f.PublicIP)
	log.Printf("-- Metadata: %s\n", f.Metadata.String())
}

func (m Map) String() string {
	output := ""
	for k, v := range m {
		output += fmt.Sprintf("(%s => %s) ", k, v)
	}
	return output
}

func CreateMasterUnits(fleetMachine *FleetMachine) []string {
	unitPathInfo := setupUnitFilesDeps()

	files := map[string]string{
		"api":        "master-apiserver@.service",
		"controller": "master-controller-manager@.service",
		"scheduler":  "master-scheduler@.service",
		"download":   "master-download-kubernetes@.service",
		"dns":        "master-dns@.service",
	}

	createdFiles := []string{}

	// Form download service file from template
	readfile, err := ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["download"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	download := string(readfile)
	download = strings.Replace(download, "<ID>", fleetMachine.ID, -1)
	download = strings.Replace(download, "<URL>", Conf.BinariesURL, -1)

	// Write download service file
	filename := strings.Replace(files["download"], "@", "@"+fleetMachine.ID, -1)
	download_file := fmt.Sprintf("%s/%s",
		unitPathInfo[0]["path"], filename)
	err = ioutil.WriteFile(download_file, []byte(download), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, download_file)

	// Form apiserver service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["api"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	apiserver := string(readfile)
	apiserver = strings.Replace(apiserver, "<ID>", fleetMachine.ID, -1)
	apiserver = strings.Replace(apiserver, "<IP_ADDR>", fleetMachine.PublicIP, -1)

	// Write apiserver service file
	filename = strings.Replace(files["api"], "@", "@"+fleetMachine.ID, -1)
	apiserver_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(apiserver_file, []byte(apiserver), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, apiserver_file)

	// Form controller service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["controller"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	controller := string(readfile)
	controller = strings.Replace(controller, "<ID>", fleetMachine.ID, -1)

	// Write controller service file
	filename = strings.Replace(files["controller"], "@", "@"+fleetMachine.ID, -1)
	controller_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(controller_file, []byte(controller), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, controller_file)

	// Form scheduler service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["scheduler"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	scheduler := string(readfile)
	scheduler = strings.Replace(scheduler, "<ID>", fleetMachine.ID, -1)

	// Write scheduler service file
	filename = strings.Replace(files["scheduler"], "@", "@"+fleetMachine.ID, -1)
	scheduler_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(scheduler_file, []byte(scheduler), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, scheduler_file)

	// Substitute the machine ID and file paths into systemd file
	filename = strings.Replace(files["dns"], "@", "@"+fleetMachine.ID, -1)
	dns_file := fmt.Sprintf("%s/%s", unitPathInfo[0]["path"], filename)
	readfile, err = ioutil.ReadFile(fmt.Sprintf("/templates/%s", files["dns"]))
	goutils.PrintErrors(goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

	dns_content := strings.Replace(string(readfile), "<NS_URL>", Conf.KubeSystemNamespace, -1)
	dns_content = strings.Replace(dns_content, "<RC_URL>", Conf.SkyDNSRepContr, -1)
	dns_content = strings.Replace(dns_content, "<SVC_URL>", Conf.SkyDNSService, -1)
	dns_content = strings.Replace(dns_content, "<ID>", fleetMachine.ID, -1)
	dns_content = strings.Replace(dns_content, "<MASTER_IP>", fleetMachine.PublicIP, -1)

	err = ioutil.WriteFile(dns_file, []byte(dns_content), 0644)
	goutils.PrintErrors(goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, dns_file)

	log.Printf("Created all unit files for: %s\n", fleetMachine.ID)
	return createdFiles
}

func CreateMinionUnits(masterFleetMachine,
	fleetMachine *FleetMachine) []string {

	masterIPPort := fmt.Sprintf("%s:%s",
		masterFleetMachine.PublicIP, Conf.KubernetesAPIPort)
	unitPathInfo := setupUnitFilesDeps()
	files := map[string]string{
		"kubelet":  "minion-kubelet@.service",
		"proxy":    "minion-proxy@.service",
		"download": "minion-download-kubernetes@.service",
	}

	createdFiles := []string{}

	// Form download service file from template
	readfile, err := ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["download"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	download := string(readfile)
	download = strings.Replace(download, "<ID>", fleetMachine.ID, -1)
	download = strings.Replace(download, "<URL>", Conf.BinariesURL, -1)

	// Write download service file
	filename := strings.Replace(files["download"], "@", "@"+fleetMachine.ID, -1)
	download_file := fmt.Sprintf("%s/%s",
		unitPathInfo[0]["path"], filename)
	err = ioutil.WriteFile(download_file, []byte(download), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, download_file)

	// Form kubelet service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["kubelet"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	kubelet := string(readfile)
	kubelet = strings.Replace(kubelet, "<ID>", fleetMachine.ID, -1)
	kubelet = strings.Replace(kubelet, "<IP_ADDR>", fleetMachine.PublicIP, -1)
	kubelet = strings.Replace(kubelet, "<MASTER_IP_PORT>", masterIPPort, -1)

	// Write kubelet service file
	filename = strings.Replace(files["kubelet"], "@", "@"+fleetMachine.ID, -1)
	kubelet_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(kubelet_file, []byte(kubelet), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, kubelet_file)

	// Form proxy service file from template
	readfile, err = ioutil.ReadFile(
		fmt.Sprintf("/templates/%s", files["proxy"]))
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	proxy := string(readfile)
	proxy = strings.Replace(proxy, "<ID>", fleetMachine.ID, -1)
	proxy = strings.Replace(proxy, "<MASTER_IP_PORT>", masterIPPort, -1)

	// Write proxy service file
	filename = strings.Replace(files["proxy"], "@", "@"+fleetMachine.ID, -1)
	proxy_file := fmt.Sprintf("%s/%s", unitPathInfo[1]["path"], filename)
	err = ioutil.WriteFile(proxy_file, []byte(proxy), 0644)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	createdFiles = append(createdFiles, proxy_file)

	log.Printf("Created all unit files for: %s\n", fleetMachine.ID)
	return createdFiles
}

func setupUnitFilesDeps() []map[string]string {
	unitPathInfo := getUnitPathInfo()
	perm := os.FileMode(os.ModeDir)

	for _, v := range unitPathInfo {
		err := os.RemoveAll(v["path"])
		goutils.PrintErrors(
			goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

		os.MkdirAll(v["path"], perm)
	}

	return unitPathInfo
	/*
		switch fleetMachine.Metadata["kubernetes_role"] {
		case "master":
			createdFiles = createMasterUnits(fleetMachine, unitPathInfo)
		case "minion":
			createdFiles = createMinionUnits(masterIPPort, fleetMachine, unitPathInfo)
		}

		log.Printf("Created all unit files for: %s\n", fleetMachine.ID)
		return createdFiles
	*/
}

func getSubStateByPath(path string) string {
	unitPathInfo := getUnitPathInfo()

	for _, v := range unitPathInfo {
		if v["path"] == path {
			return v["subState"]
		}
	}

	return ""
}

func getUnitPathInfo() []map[string]string {
	templatePath := "/units/kubernetes_units"
	unitPathInfo := []map[string]string{}

	unitPathInfo = append(unitPathInfo,
		map[string]string{
			"path":     templatePath + "/download",
			"subState": "exited",
		},
	)

	unitPathInfo = append(unitPathInfo,
		map[string]string{
			"path":     templatePath + "/roles",
			"subState": "running",
		},
	)

	return unitPathInfo
}

func getUnitState(unitFile string) FleetUnitState {
	var fleetUnitStates FleetUnitStates
	filename := filepath.Base(unitFile)

	urlPath := fmt.Sprintf("fleet/%s/state", Conf.FleetAPIVersion)
	url := getFullAPIURL(Conf.FleetAPIPort, urlPath)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             url,
		Headers:         headers,
	}

	_, jsonResponse, _ := goutils.HttpCreateRequest(p)
	err := json.Unmarshal(jsonResponse, &fleetUnitStates)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

	for _, unitState := range fleetUnitStates.States {
		if unitState.Name == filename {
			return unitState
		}
	}
	return FleetUnitState{}
}

func lowerCasingOfUnitOptionsStr(json_str string) string {
	json_str = strings.Replace(json_str, "\"Section\"", "\"section\"", -1)
	json_str = strings.Replace(json_str, "\"Name\"", "\"name\"", -1)
	json_str = strings.Replace(json_str, "\"Value\"", "\"value\"", -1)

	return json_str
}

func StartUnitFile(unitFile string) {
	filename := filepath.Base(unitFile)
	unitFilepath := fmt.Sprintf(
		"fleet/%s/units/%s", Conf.FleetAPIVersion, filename)
	url := getFullAPIURL(Conf.FleetAPIPort, unitFilepath)

	log.Printf("Starting unit file: %s", filename)

	statusCode := 0
	for statusCode != 204 {
		readfile, err := ioutil.ReadFile(unitFile)
		goutils.PrintErrors(
			goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

		content := string(readfile)
		u, _ := unit.NewUnitFile(content)

		options_bytes, _ := json.Marshal(u.Options)
		options_str := lowerCasingOfUnitOptionsStr(string(options_bytes))

		json_str := fmt.Sprintf(
			`{"name": "%s", "desiredState":"launched", "options": %s}`,
			filename,
			options_str)

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		p := goutils.HttpRequestParams{
			HttpRequestType: "PUT",
			Url:             url,
			Data:            json_str,
			Headers:         headers,
		}
		statusCode, _, _ = goutils.HttpCreateRequest(p)

		time.Sleep(1 * time.Second)
		/*
			log.Printf(
				"curl -H \"Content-Type: application/json\" -X PUT "+
					"-d %q localhost:10001/v1-alpha/units/%s",
				json_str, filename)
			body, err := ioutil.ReadAll(resp.Body)
			log.Printf("Status Code: %s", statusCode)
			log.Printf("[Error] in HTTP Body: %s - %v", body, err)
			goutils.PrintErrors(
				goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
		*/
	}
}

func UnitFileCompleted(unitFile string) bool {
	filename := filepath.Base(unitFile)
	dir := filepath.Dir(unitFile)

	expectedSubState := getSubStateByPath(dir)
	unitState := getUnitState(unitFile)

	if unitState.Name == filename {
		if unitState.SystemdSubState == expectedSubState {
			return true
		}
	}

	return false
}

func WaitUnitFileComplete(unitFile string) {
	filename := filepath.Base(unitFile)

	complete := false
	for !complete {
		complete = UnitFileCompleted(unitFile)

		if !complete {
			log.Printf("-- Waiting for the following unit file to complete: %s",
				filename)
			time.Sleep(1 * time.Second)
		}
	}

	log.Printf("The following unit file has completed: %s", filename)
}
