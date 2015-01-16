package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"time"

	"github.com/metral/corekube_travis"
	"github.com/metral/goheat"
	"github.com/metral/goheat/rax"
	"github.com/metral/goheat/util"
	"github.com/metral/goutils"
	"github.com/metral/overlord/lib"
)

func overlayNetworksCountTest(config *util.HeatConfig, details *util.StackDetails) string {
	d := *details
	msg := ""
	sleepDuration := 10 //seconds

	for {
		msg = "overlayNetworksCountTest: "

		overlordIP := util.ExtractOverlordIP(d)
		masterCount, _ := strconv.Atoi(
			d.Stack.Parameters["kubernetes-master-count"].(string))
		minionCount, _ := strconv.Atoi(
			d.Stack.Parameters["kubernetes-minion-count"].(string))
		expectedCount := masterCount + minionCount

		var overlayResult lib.Result
		path := fmt.Sprintf("%s/keys/coreos.com/network/subnets",
			lib.ETCD_API_VERSION)
		url := fmt.Sprintf("http://%s:%s/%s",
			overlordIP, lib.ETCD_CLIENT_PORT, path)

		token := rax.IdentitySetup(config)

		headers := map[string]string{
			"X-Auth-Token": token.ID,
			"Content-Type": "application/json",
		}

		p := goutils.HttpRequestParams{
			HttpRequestType: "GET",
			Url:             url,
			Headers:         headers,
		}

		_, jsonResponse := goutils.HttpCreateRequest(p)
		err := json.Unmarshal(jsonResponse, &overlayResult)
		goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 2})

		overlayNetworksCount := len(overlayResult.Node.Nodes)

		if overlayNetworksCount == expectedCount {
			return "Passed"
		}

		msg += fmt.Sprintf("ExpectedCount: %d, OverlayNetworkCount: %d",
			expectedCount, overlayNetworksCount)
		log.Printf(msg)
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}

	return "Failed"
}

func runTests(config *util.HeatConfig, details *util.StackDetails) {
	corekube_travis.StartTestTimeout(1, config, details, overlayNetworksCountTest)
}

func main() {
	config, stackDetails := corekube_travis.BuildConfigAndCreateStack()
	runTests(config, stackDetails)
	goheat.DeleteStack(config, stackDetails.Stack.Links[0].Href)
}
