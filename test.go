package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/metral/corekube-travis/rax"
	"github.com/metral/goutils"
)

type HeatStack struct {
	Name            string            `json:"stack_name"`
	Template        string            `json:"template"`
	Params          map[string]string `json:"parameters"`
	Timeout         int               `json:"timeout_mins"`
	DisableRollback bool              `json:"disable_rollback"`
}

type CreateStackResult struct {
	Stack CreateStackResultData `json:"stack"`
}

type CreateStackResultData struct {
	Id    string       `json:"id"`
	Links []StackLinks `json:"links"`
}

type StackLinks struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
}

type StackDetails struct {
	Stack StackDetailsData `json:"stack"`
}

type StackDetailsData struct {
	StackStatus       string               `json:"stack_status"`
	StackStatusReason string               `json:"stack_status_reason"`
	Id                string               `json:"id"`
	Outputs           []StackDetailsOutput `json:"outputs"`
}

type StackDetailsOutput struct {
	OutputValue interface{} `json:"output_value"`
	Description string      `json:"description"`
	OutputKey   string      `json:"output_key"`
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

func getStackDetails(result *CreateStackResult) StackDetails {
	var details StackDetails
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

func watchStackCreation(result *CreateStackResult) StackDetails {
	sleepDuration := 10 // seconds
	var details StackDetails

watchLoop:
	for {
		details = getStackDetails(result)
		log.Printf("Stack Status: %s", details.Stack.StackStatus)

		switch details.Stack.StackStatus {
		case "CREATE_IN_PROGRESS":
			time.Sleep(time.Duration(sleepDuration) * time.Second)
		default:
			break watchLoop
		}
	}

	return details
}

func startStackTimeout(heatTimeout int, result *CreateStackResult) StackDetails {
	chan1 := make(chan StackDetails, 1)
	go func() {
		stackDetails := watchStackCreation(result)
		chan1 <- stackDetails
	}()

	select {
	case result := <-chan1:
		return result
	case <-time.After(time.Duration(heatTimeout) * time.Minute):
		msg := fmt.Sprintf("Stack create timed out after %d mins", heatTimeout)
		log.Fatal(msg)
	}
	return *new(StackDetails)
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

	s := &HeatStack{
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

func createStack(templateFile, keyName string) CreateStackResult {
	readfile, _ := ioutil.ReadFile(templateFile)
	template := string(readfile)
	var result CreateStackResult

	token := rax.IdentitySetup()

	statusCode, bodyBytes := createStackReq(template, token.ID, keyName)

	switch statusCode {
	case 201:
		err := json.Unmarshal(bodyBytes, &result)
		goutils.CheckForErrors(goutils.ErrorParams{Err: err, CallerNum: 1})
	}
	return result
}

func testMinionsRegistered(machines []string, k8sTimeout int) {
	log.Printf("%s", machines)
}

func main() {
	heatTimeout := 10 // minutes
	//k8sTimeout := 1   // minutes
	templateFile := "../../../corekube-heat.yaml"
	keyName := "argon_dfw"

	result := createStack(templateFile, keyName)
	stackDetails := startStackTimeout(heatTimeout, &result)
	for _, i := range stackDetails.Stack.Outputs {
		log.Printf("%s", i)
	}
	//testMinionsRegistered(machines, k8sTimeout)
}
