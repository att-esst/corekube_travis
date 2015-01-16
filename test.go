package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strconv"

	"time"

	"github.com/metral/goheat"
	"github.com/metral/goheat/rax"
	"github.com/metral/goheat/util"
	"github.com/metral/goutils"
	"github.com/metral/overlord/lib"
)

var (
	templateFile = flag.String("templateFile", "", "Path of corekube-heat.yaml")
	keypair      = flag.String("keypair", "", "Existing SSH keypair")
	authUrl      = flag.String("authUrl", "", "Openstack Auth URL")
	username     = flag.String("username", "", "Openstack Username")
	password     = flag.String("password", "", "Openstack Password")
	tenantId     = flag.String("tenantId", "", "Openstack Tenant ID")
	timeout      = flag.Int("timeout", 10, "Heat stack create timeout")
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

func startTestTimeout(timeout int, config *util.HeatConfig, details *util.StackDetails, f func(*util.HeatConfig, *util.StackDetails) string) {
	chan1 := make(chan string, 1)
	go func() {
		result := f(config, details)
		chan1 <- result
	}()

	select {
	case result := <-chan1:
		msg := fmt.Sprintf("%s %s.", util.GetFunctionName(f), result)
		log.Printf(msg)
	case <-time.After(time.Duration(timeout) * time.Minute):
		msg := fmt.Sprintf("%s Failed: timed out after %d mins.",
			util.GetFunctionName(f), timeout)
		log.Fatal(msg)
	}
}

func runTests(config *util.HeatConfig, details *util.StackDetails) {
	startTestTimeout(1, config, details, overlayNetworksCountTest)
}

func main() {
	flag.Parse()

	c := util.HeatConfig{
		TemplateFile: *templateFile,
		Keypair:      *keypair,
		OSAuthUrl:    *authUrl,
		OSUsername:   *username,
		OSPassword:   *password,
		OSTenantId:   *tenantId,
		Timeout:      int(*timeout),
	}

	result := goheat.CreateStack(&c)
	stackDetails := goheat.StartStackTimeout(&c, &result)
	runTests(&c, &stackDetails)
	deleteStack(stackDetails.Stack.Links[0].Href)
}
