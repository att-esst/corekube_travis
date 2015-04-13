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
}

var Conf = new(VersionsConf)

func init() {
	file, _ := os.Open("/conf.json")
	json.NewDecoder(file).Decode(Conf)
	file.Close()
}
