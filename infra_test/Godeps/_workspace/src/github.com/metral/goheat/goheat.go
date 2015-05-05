package goheat

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"time"

	"github.com/metral/goheat/rax"
	"github.com/metral/goheat/util"
	"github.com/metral/goutils"
)

var (
	templateFilepath = flag.String("templateFilePath", "", "Filepath of corekube-heat.yaml")
)

func getStackDetails(config *util.HeatConfig, result *util.CreateStackResult) util.StackDetails {
	var details util.StackDetails
	url := (*result).Stack.Links[0].Href
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

	statusCode, bodyBytes, _ := goutils.HttpCreateRequest(p)

	switch statusCode {
	case 200:
		err := json.Unmarshal(bodyBytes, &details)
		goutils.PrintErrors(
			goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	}

	return details
}

func watchStackCreation(config *util.HeatConfig, result *util.CreateStackResult) util.StackDetails {
	sleepDuration := 10 // seconds
	var details util.StackDetails

watchLoop:
	for {
		details = getStackDetails(config, result)
		log.Printf("Stack Status: %s", details.Stack.StackStatus)

		switch details.Stack.StackStatus {
		case "CREATE_IN_PROGRESS":
			time.Sleep(time.Duration(sleepDuration) * time.Second)
		case "CREATE_COMPLETE":
			break watchLoop
		default:
			log.Printf("Stack Status: %s", details.Stack.StackStatus)
			log.Printf("Stack Status: %s", details.Stack.StackStatusReason)
			DeleteStack(config, result.Stack.Links[0].Href)
			log.Fatal()
		}
	}

	return details
}

func StartStackTimeout(config *util.HeatConfig, result *util.CreateStackResult) util.StackDetails {
	chan1 := make(chan util.StackDetails, 1)
	go func() {
		stackDetails := watchStackCreation(config, result)
		chan1 <- stackDetails
	}()

	select {
	case result := <-chan1:
		return result
	case <-time.After(time.Duration(config.Timeout) * time.Minute):
		msg := fmt.Sprintf("Stack create timed out after %d mins", config.Timeout)
		DeleteStack(config, result.Stack.Links[0].Href)
		log.Fatal(msg)
	}
	return *new(util.StackDetails)
}

func createStackReq(
	template, token, keypair string, extraParams *map[string]string) (int, []byte) {

	timeout := int(10)
	params := map[string]string{
		"keyname": keypair,
	}

	if len(*extraParams) > 0 {
		for k, v := range *extraParams {
			params[k] = v
		}
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

	statusCode, bodyBytes, _ := goutils.HttpCreateRequest(h)
	return statusCode, bodyBytes
}

func CreateStack(
	params *map[string]string, config *util.HeatConfig) util.CreateStackResult {

	readfile, _ := ioutil.ReadFile(config.TemplateFile)
	template := string(readfile)
	var result util.CreateStackResult

	token := rax.IdentitySetup(config)

	statusCode, bodyBytes := createStackReq(
		template, token.ID, config.Keypair, params)

	switch statusCode {
	case 201:
		err := json.Unmarshal(bodyBytes, &result)
		goutils.PrintErrors(
			goutils.ErrorParams{Err: err, CallerNum: 2, Fatal: false})
	}
	return result
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

func DeleteStack(config *util.HeatConfig, stackUrl string) {
	token := rax.IdentitySetup(config)

	headers := map[string]string{
		"X-Auth-Token": token.ID,
		"Content-Type": "application/json",
	}

	p := goutils.HttpRequestParams{
		HttpRequestType: "DELETE",
		Url:             stackUrl,
		Headers:         headers,
	}

	statusCode, _, _ := goutils.HttpCreateRequest(p)

	switch statusCode {
	case 204:
		log.Printf("Delete stack requested.")
	}

}
