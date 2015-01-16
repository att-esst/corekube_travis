package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/metral/corekube_travis"
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

func minionK8sCountTest(
	config *util.HeatConfig, details *util.StackDetails) string {

	d := *details
	msg := ""
	sleepDuration := 10 //seconds

	for {
		msg = "minionK8sCountTest: "

		masterIP := util.ExtractArrayIPs(d, "master_ips")
		//minionCount, _ := strconv.Atoi(
		//		d.Stack.Parameters["kubernetes-minion-count"].(string))

		//var minionsResult lib.Result
		endpoint := fmt.Sprintf("http://%s:%s", masterIP[0], lib.K8S_API_PORT)
		masterAPIurl := fmt.Sprintf(
			"%s/api/%s/minions", endpoint, lib.K8S_API_VERSION)

		headers := map[string]string{
			"Content-Type": "application/json",
		}

		p := goutils.HttpRequestParams{
			HttpRequestType: "GET",
			Url:             masterAPIurl,
			Headers:         headers,
		}

		_, jsonResponse := goutils.HttpCreateRequest(p)
		//err := json.Unmarshal(jsonResponse, &minionsResult)
		log.Printf("jsonreponse: %s", jsonResponse)
		/*
			goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 2})

			overlayNetworksCount := len(overlayResult.Node.Nodes)

			if overlayNetworksCount == expectedCount {
				return "Passed"
			}

			msg += fmt.Sprintf("ExpectedCount: %d, OverlayNetworkCount: %d",
				expectedCount, overlayNetworksCount)
		*/
		log.Printf(msg)
		time.Sleep(time.Duration(sleepDuration) * time.Second)
	}

	return "Failed"
}

func runTests(config *util.HeatConfig, details *util.StackDetails) {
	corekube_travis.StartTestTimeout(2, config, details, minionK8sCountTest)
}

func main() {
	params := map[string]string{
		"git-command": createGitCmdParam(),
	}
	config, stackDetails := corekube_travis.BuildConfigAndCreateStack(&params)
	runTests(config, stackDetails)
	//goheat.DeleteStack(config, stackDetails.Stack.Links[0].Href)
}
