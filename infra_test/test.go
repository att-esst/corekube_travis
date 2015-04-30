package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"time"

	"github.com/metral/corekube_travis/framework"
	"github.com/metral/goheat"
	"github.com/metral/goheat/rax"
	"github.com/metral/goheat/util"
	"github.com/metral/goutils"
	"github.com/metral/overlord/lib"
)

func overlayNetworksCountTest(
	config *util.HeatConfig, details *util.StackDetails) string {

	d := *details
	msg := ""
	sleepDuration := 10 //seconds

	for {
		msg = "overlayNetworksCountTest: "

		overlordIP := util.ExtractIPFromStackDetails(d, "overlord_ip")
		masterCount := 1
		minionCount, _ := strconv.Atoi(
			d.Stack.Parameters["kubernetes_minion_count"].(string))
		expectedCount := masterCount + minionCount

		var overlayResult lib.Result
		path := fmt.Sprintf("%s/keys/coreos.com/network/subnets",
			lib.Conf.EtcdAPIVersion)
		url := fmt.Sprintf("http://%s:%s/%s",
			overlordIP, lib.Conf.EtcdClientPort, path)

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

		_, jsonResponse, createError := goutils.HttpCreateRequest(p)
		goutils.PrintErrors(
			goutils.ErrorParams{Err: createError, CallerNum: 2, Fatal: false})
		err := json.Unmarshal(jsonResponse, &overlayResult)
		goutils.PrintErrors(
			goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})

		overlayNetworksCount := len(overlayResult.Node.Nodes)

		msg += fmt.Sprintf("ExpectedCount: %d, OverlayNetworkCount: %d",
			expectedCount, overlayNetworksCount)
		log.Printf(msg)

		if overlayNetworksCount == expectedCount {
			return "Passed"
		}

		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}

	return "Failed"
}

func runTests(config *util.HeatConfig, details *util.StackDetails) {
	framework.StartTestTimeout(10, config, details, overlayNetworksCountTest)
}

func main() {
	params := map[string]string{}
	config, stackDetails := framework.BuildConfigAndCreateStack(&params)
	runTests(config, stackDetails)
	goheat.DeleteStack(config, stackDetails.Stack.Links[0].Href)
}
