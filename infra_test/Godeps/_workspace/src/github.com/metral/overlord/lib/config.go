package lib

import (
	"encoding/json"
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
	KubeSystemNamespace  string `json:"kubeSystemNS"`
	SkyDNSService        string `json:"skyDNSSVC"`
	SkyDNSRepContr       string `json:"skyDNSRC"`
}

var Conf = new(VersionsConf)

var (
	conf_file = "/tmp/conf.json"
)

func init() {
	file, _ := os.Open(conf_file)
	json.NewDecoder(file).Decode(Conf)

	file.Close()
}
