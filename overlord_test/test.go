package main

import (
	"github.com/metral/corekube_travis"
	"github.com/metral/goheat"
	"github.com/metral/goheat/util"
)

func runTests(config *util.HeatConfig, details *util.StackDetails) {
	//corekube_travis.StartTestTimeout(1, config, details, overlayNetworksCountTest)
}

func main() {
	params := map[string]string{}
	config, stackDetails := corekube_travis.BuildConfigAndCreateStack(&params)
	runTests(config, stackDetails)
	goheat.DeleteStack(config, stackDetails.Stack.Links[0].Href)
}
