package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/metral/corekube-travis/rax"
	"github.com/metral/corekube-travis/util"
	"github.com/metral/goutils"
	"github.com/metral/overlord/lib"
)

var (
	templateFilepath = flag.String("templateFilePath", "", "Filepath of corekube-heat.yaml")
)

func createGitCmdParam() string {
	travisPR := os.Getenv("TRAVIS_PULL_REQUEST")
	travisRepoSlug := os.Getenv("TRAVIS_REPO_SLUG")

	repo := fmt.Sprintf("https://github.com/%s", travisRepoSlug)
	cmd := ""

	switch travisPR {
	case "false": // false aka build commit
		travisBranch := os.Getenv("TRAVIS_BRANCH")
		travisCommit := os.Getenv("TRAVIS_COMMIT")
		c := []string{
			fmt.Sprintf("/usr/bin/git clone -b %s %s", travisBranch, repo),
			fmt.Sprintf("/usr/bin/git checkout -qf %s", travisCommit),
		}
		cmd = strings.Join(c, "; ")
	default: // PR number
		c := []string{
			fmt.Sprintf("/usr/bin/git clone %s", repo),
			fmt.Sprintf("/usr/bin/git fetch origin +refs/pull/%s/merge",
				travisPR),
			fmt.Sprintf("/usr/bin/git checkout -qf FETCH_HEAD"),
		}
		cmd = strings.Join(c, "; ")
	}

	return cmd
}

func getStackDetails(result *util.CreateStackResult) util.StackDetails {
	var details util.StackDetails
	url := (*result).Stack.Links[0].Href
	token := rax.IdentitySetup()

	headers := map[string]string{
		"X-Auth-Token": token.ID,
		"Content-Type": "application/json",
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "GET",
		Url:             url,
		Headers:         headers,
	}

	statusCode, bodyBytes := goutils.HttpCreateRequest(p)

	switch statusCode {
	case 200:
		err := json.Unmarshal(bodyBytes, &details)
		goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})
	}

	return details
}

func watchStackCreation(result *util.CreateStackResult) util.StackDetails {
	sleepDuration := 10 // seconds
	var details util.StackDetails

watchLoop:
	for {
		details = getStackDetails(result)
		log.Printf("Stack Status: %s", details.Stack.StackStatus)

		switch details.Stack.StackStatus {
		case "CREATE_IN_PROGRESS":
			time.Sleep(time.Duration(sleepDuration) * time.Second)
		case "CREATE_COMPLETE":
			break watchLoop
		default:
			log.Printf("Stack Status: %s", details.Stack.StackStatus)
			log.Printf("Stack Status: %s", details.Stack.StackStatusReason)
			deleteStack(result.Stack.Links[0].Href)
			log.Fatal()
		}
	}

	return details
}

func startStackTimeout(heatTimeout int, result *util.CreateStackResult) util.StackDetails {
	chan1 := make(chan util.StackDetails, 1)
	go func() {
		stackDetails := watchStackCreation(result)
		chan1 <- stackDetails
	}()

	select {
	case result := <-chan1:
		return result
	case <-time.After(time.Duration(heatTimeout) * time.Minute):
		msg := fmt.Sprintf("Stack create timed out after %d mins", heatTimeout)
		deleteStack(result.Stack.Links[0].Href)
		log.Fatal(msg)
	}
	return *new(util.StackDetails)
}

func createStackReq(template, token, keyName string) (int, []byte) {
	timeout := int(10)
	params := map[string]string{
		"git-command": createGitCmdParam(),
		"key-name":    keyName,
	}
	disableRollback := bool(false)

	timestamp := int32(time.Now().Unix())
	templateName := fmt.Sprintf("corekube-travis-%d", timestamp)
	log.Printf("Started creating stack: %s", templateName)

	s := &util.HeatStack{
		Name:            templateName,
		Template:        template,
		Params:          params,
		Timeout:         timeout,
		DisableRollback: disableRollback,
	}
	jsonByte, _ := json.Marshal(s)

	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Auth-Token": token,
	}

	urlStr := fmt.Sprintf("%s/stacks", os.Getenv("TRAVIS_OS_HEAT_URL"))

	h := goutils.HttpRequestParams{
		HttpRequestType: "POST",
		Url:             urlStr,
		Data:            jsonByte,
		Headers:         headers,
	}

	statusCode, bodyBytes := goutils.HttpCreateRequest(h)
	return statusCode, bodyBytes
}

func createStack(templateFile, keyName string) util.CreateStackResult {
	readfile, _ := ioutil.ReadFile(templateFile)
	template := string(readfile)
	var result util.CreateStackResult

	token := rax.IdentitySetup()

	statusCode, bodyBytes := createStackReq(template, token.ID, keyName)

	switch statusCode {
	case 201:
		err := json.Unmarshal(bodyBytes, &result)
		goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})
	}
	return result
}

func overlayNetworksCountTest(details *util.StackDetails) {
	d := *details

	overlordIP := extractOverlordIP(d)
	masterCount, _ := strconv.Atoi(
		d.Stack.Parameters["kubernetes-master-count"].(string))
	minionCount, _ := strconv.Atoi(
		d.Stack.Parameters["kubernetes-minion-count"].(string))
	totalCount := masterCount + minionCount

	var subnetResult lib.Result
	path := fmt.Sprintf("%s/keys/coreos.com/network/subnets",
		lib.ETCD_API_VERSION)
	url := fmt.Sprintf("http://%s:%s/%s", overlordIP, lib.ETCD_CLIENT_PORT, path)

	token := rax.IdentitySetup()

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
	err := json.Unmarshal(jsonResponse, &subnetResult)
	goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 2})

	subnetCount := len(subnetResult.Node.Nodes)

	if subnetCount != totalCount {
		msg := fmt.Sprintf("Test Failed: overlayNetworksCountTest:"+
			" ExpectedCount: %d, OverlayNetworkCount: %d",
			totalCount, subnetCount)
		deleteStack(details.Stack.Links[0].Href)
		log.Fatal(msg)
	}
	log.Printf("Test Succeeded: overlayNetworksCountTest")
}

func runTests(details *util.StackDetails) {
	overlayNetworksCountTest(details)
}

func extractOverlordIP(details util.StackDetails) string {
	overlordIP := ""

	for _, i := range details.Stack.Outputs {
		if i.OutputKey == "overlord_ip" {
			overlordIP = i.OutputValue.(string)
		}
	}

	return overlordIP
}

func deleteStack(stackUrl string) {
	token := rax.IdentitySetup()

	headers := map[string]string{
		"X-Auth-Token": token.ID,
		"Content-Type": "application/json",
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "DELETE",
		Url:             stackUrl,
		Headers:         headers,
	}

	statusCode, _ := goutils.HttpCreateRequest(p)

	switch statusCode {
	case 204:
		log.Printf("Delete stack requested.")
	}

}

func main() {
	flag.Parse()

	heatTimeout := 10 // minutes
	templateFile := *templateFilepath
	keyName := os.Getenv("TRAVIS_OS_KEYPAIR")

	result := createStack(templateFile, keyName)
	stackDetails := startStackTimeout(heatTimeout, &result)
	runTests(&stackDetails)
	deleteStack(stackDetails.Stack.Links[0].Href)
}
