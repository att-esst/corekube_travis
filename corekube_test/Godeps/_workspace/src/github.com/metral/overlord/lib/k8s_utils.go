package lib

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1beta3"
	"github.com/metral/goutils"
)

type PreregisteredKNode struct {
	Kind       string                 `json:"kind,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Status     v1beta3.NodeStatus     `json:"status,omitempty"`
	APIVersion string                 `json:"apiVersion,omitempty"`
}

type KNodesResult struct {
	Kind              string `json:"kind,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
	SelfLink          string `json:"selfLink,omitempty"`
	APIVersion        string `json:"apiVersion,omitempty"`
	Nodes             KNodes `json:"nodes,omitempty"`
}

type KNodesCountResult struct {
	Kind              string `json:"kind,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
	SelfLink          string `json:"selfLink,omitempty"`
	APIVersion        string `json:"apiVersion,omitempty"`
	Items             KNodes `json:"items,omitempty"`
}

type KNodes []KNode
type KNode struct {
	ID                string `json:"id,omitempty"`
	UID               string `json:"uid,omitempty"`
	CreationTimestamp string `json:"creationTimestamp,omitempty"`
	SelfLink          string `json:"selfLink,omitempty"`
	ResourceVersion   int    `json:"resourceVersion,omitempty"`
	HostIP            string `json:"hostIP,omitempty"`
	Resources         map[interface{}]interface{}
}

func isMaster(fleetMachine *FleetMachine) bool {
	role := fleetMachine.Metadata["kubernetes_role"]

	switch role {
	case "master":
		return true
	}
	return false
}

func isMinion(fleetMachine *FleetMachine) bool {
	role := fleetMachine.Metadata["kubernetes_role"]

	switch role {
	case "minion":
		return true
	}
	return false
}

func registerKNodes(master *FleetMachine, node *FleetMachine) {

	// Get registered nodes, if any
	endpoint := fmt.Sprintf("http://%s:%s", master.PublicIP, Conf.KubernetesAPIPort)
	masterAPIurl := fmt.Sprintf("%s/api/%s/nodes", endpoint, Conf.KubernetesAPIVersion)

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             masterAPIurl,
		Headers:         headers,
	}

	_, jsonResponse, _ := goutils.HttpCreateRequest(p)

	var nodesResult KNodesResult
	err := json.Unmarshal(jsonResponse, &nodesResult)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

	// See if nodes discovered have been registered. If not, register
	registered := false
	/*
		for _, registeredKNode := range nodesResult.Nodes {
			if registeredKNode.HostIP == node.PublicIP {
				registered = true
			}
		}
	*/

	if !registered {
		register(endpoint, node.PublicIP)
		time.Sleep(500 * time.Millisecond)
	}
}

func register(endpoint, addr string) error {
	status := v1beta3.NodeStatus{}
	status.Addresses = []v1beta3.NodeAddress{
		v1beta3.NodeAddress{
			Address: addr,
			Type:    v1beta3.NodeInternalIP,
		},
	}

	m := &PreregisteredKNode{
		Kind:       "Node",
		Metadata:   map[string]interface{}{"name": addr},
		Status:     status,
		APIVersion: Conf.KubernetesAPIVersion,
	}
	data, err := json.Marshal(m)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 1, Fatal: false})

	url := fmt.Sprintf("%s/api/%s/nodes", endpoint, Conf.KubernetesAPIVersion)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "POST",
		Url:             url,
		Data:            data,
		Headers:         headers,
	}

	statusCode, body, err := goutils.HttpCreateRequest(p)

	switch statusCode {
	case 200, 201, 202:
		log.Printf("Registered node with the Kubernetes master: %s\n", addr)
		return nil
	}
	log.Printf("%d\n%s", statusCode, body)
	goutils.PrintErrors(
		goutils.ErrorParams{Err: err, CallerNum: 1, Fatal: false})
	return nil
}
