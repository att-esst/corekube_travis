package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"
	"travis/rax"

	"github.com/metral/goutils"
)

type HeatStack struct {
	Name            string            `json:"stack_name"`
	Template        string            `json:"template"`
	Params          map[string]string `json:"parameters"`
	Timeout         int               `json:"timeout_mins"`
	DisableRollback bool              `json:"disable_rollback"`
}

func waitForMachines() {
	// - return body in goutils httpcreaterequest and adjust fleet, k8s & etcd
	// files accordingly
	// - store stack_id from response body

	// - do http GET request to stack details
	// - loop every 10 sec to get status of stack creation
	// - status can be IN-PROGRESS, FAILED, COMPLETE
	// - urlStr := fmt.Sprintf("%s/stacks", os.Getenv("TRAVIS_OS_HEAT_URL"))
}

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

func createStackReq(template, token, keyName string) {
	timeout := int(10)
	params := map[string]string{
		"git-command": createGitCmdParam(),
		"key-name":    keyName,
	}
	disableRollback := bool(false)

	s := &HeatStack{
		Name:            "corekube-travis",
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

	goutils.HttpCreateRequest(h)
}

func deployStack(templateFile, keyName string) {
	readfile, _ := ioutil.ReadFile(templateFile)
	template := string(readfile)

	token := rax.IdentitySetup()
	createStackReq(template, token.ID, keyName)
}

func waitForStackResult(heatTimeout int) []string {
	chan1 := make(chan []string, 1)
	go func() {
		machines := waitForMachines()
		chan1 <- machines
	}()

	select {
	case result := <-chan1:
		return result
	case <-time.After(time.Duration(heatTimeout) * time.Minute):
		log.Fatal("Timed out: Waiting for Heat Stack Result")
	}
	return nil
}

func testMinionsRegistered(machines []string, k8sTimeout int) {
	log.Printf("%s", machines)
}

func main() {
	//heatTimeout := 10 // minutes
	//k8sTimeout := 1   // minutes
	templateFile := "../corekube-heat.yaml"
	keyName := "argon_dfw"

	deployStack(templateFile, keyName)
	//machines := waitForStackResult(heatTimeout)
	//testMinionsRegistered(machines, k8sTimeout)
}
