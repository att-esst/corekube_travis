package lib

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

type VersionsConf struct {
	BinariesURL          string `json:"binariesURL"`
	KubernetesAPIVersion string `json:"kubernetesAPIVersion"`
	KubernetesAPIPort    string `json:"kubernetesAPIPort"`
	FleetAPIVersion      string `json:"fleetAPIVersion"`
	FleetAPIPort         string `json:"fleetAPIPort"`
	EtcdAPIVersion       string `json:"etcdAPIVersion"`
	EtcdClientPort       string `json:"etcdClientPort"`
}

var Conf = new(VersionsConf)

var (
	conf_file = "/tmp/conf.json"
	//conf_file = flag.String("conf_file", "/etc/overlord/conf.json", ""+
	//	"The conf.json file listing all versions used as well "+
	//	"as the URL of the Kubernetes binaries to deploy.")
)

func init() {
	flag.Parse()

	log.Printf("conf_file: %s", conf_file)
	file, _ := os.Open(*conf_file)
	json.NewDecoder(file).Decode(Conf)

	file.Close()
}
