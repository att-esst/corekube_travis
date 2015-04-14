package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/metral/corekube_travis"
	"github.com/metral/goheat"
	"github.com/metral/goheat/util"
	"github.com/metral/goutils"
	"github.com/metral/overlord/lib"
)

func createGitCmdParam() string {
	travisPR := os.Getenv("TRAVIS_PULL_REQUEST")
	overlordRepoSlug := "metral/overlord"

	repoURL := fmt.Sprintf("https://github.com/%s", overlordRepoSlug)
	repo := strings.Split(overlordRepoSlug, "/")[1]
	cmd := ""

	switch travisPR {
	case "false": // false aka build commit
		travisBranch := os.Getenv("TRAVIS_BRANCH")
		travisCommit := os.Getenv("TRAVIS_COMMIT")
		c := []string{
			fmt.Sprintf("/usr/bin/git clone -b %s %s", travisBranch, repoURL),
			fmt.Sprintf("/usr/bin/git -C %s checkout -qf %s", repo, travisCommit),
		}
		cmd = strings.Join(c, " ; ")
	default: // PR number
		c := []string{
			fmt.Sprintf("/usr/bin/git clone %s", repoURL),
			fmt.Sprintf("/usr/bin/git -C %s fetch origin +refs/pull/%s/merge",
				repo, travisPR),
			fmt.Sprintf("/usr/bin/git -C %s checkout -qf FETCH_HEAD", repo),
		}
		cmd = strings.Join(c, " ; ")
	}

	return cmd
}

func nodeK8sCountTest(
	config *util.HeatConfig, details *util.StackDetails) string {

	d := *details
	msg := ""
	sleepDuration := 10 //seconds

	for {
		msg = "nodeK8sCountTest: "

		masterIP := util.ExtractArrayIPs(d, "master_ips")
		expectedNodeCount, _ := strconv.Atoi(
			d.Stack.Parameters["kubernetes-minion-count"].(string))

		var nodesResult lib.KNodesCountResult
		endpoint := fmt.Sprintf("http://%s:%s", masterIP[0], lib.Conf.KubernetesAPIPort)
		masterAPIurl := fmt.Sprintf(
			"%s/api/%s/nodes", endpoint, lib.Conf.KubernetesAPIVersion)
		log.Printf("url :%s", masterAPIurl)

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		p := goutils.HttpRequestParams{
			HttpRequestType: "GET",
			Url:             masterAPIurl,
			Headers:         headers,
		}

		_, bodyBytes, _ := goutils.HttpCreateRequest(p)

		json.Unmarshal(bodyBytes, &nodesResult)
		nodesCount := len(nodesResult.Items)

		for x := range nodesResult.Items {
			log.Printf("k8s node: %v", x)
		}

		msg += fmt.Sprintf("ExpectedCount: %d, NodeCount: %d",
			expectedNodeCount, nodesCount)
		log.Printf(msg)

		if nodesCount == expectedNodeCount {
			return "Passed"
		}

		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}

	return "Failed"
}

func runTests(config *util.HeatConfig, details *util.StackDetails) {
	corekube_travis.StartTestTimeout(10, config, details, nodeK8sCountTest)
}

func main() {
	params := map[string]string{
		"git-command": createGitCmdParam(),
	}
	config, stackDetails := corekube_travis.BuildConfigAndCreateStack(&params)
	runTests(config, stackDetails)
	goheat.DeleteStack(config, stackDetails.Stack.Links[0].Href)
}
