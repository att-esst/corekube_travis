package framework

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/metral/goheat"
	"github.com/metral/goheat/util"
)

var (
	templateFile = flag.String("templateFile", "", "Path of corekube Heat template")
	keypair      = flag.String("keypair", "", "Existing SSH keypair")
	authUrl      = flag.String("authUrl", "", "Openstack Auth URL")
	username     = flag.String("username", "", "Openstack Username")
	password     = flag.String("password", "", "Openstack Password")
	tenantId     = flag.String("tenantId", "", "Openstack Tenant ID")
	timeout      = flag.Int("timeout", 10, "Heat stack create timeout")
)

func StartTestTimeout(timeout int,
	config *util.HeatConfig,
	details *util.StackDetails,
	f func(*util.HeatConfig, *util.StackDetails) string) {

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

func BuildConfigAndCreateStack(
	params *map[string]string) (*util.HeatConfig, *util.StackDetails) {

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

	result := goheat.CreateStack(params, &c)
	stackDetails := goheat.StartStackTimeout(&c, &result)

	return &c, &stackDetails
}
