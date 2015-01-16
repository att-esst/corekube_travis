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
	config, stackDetails := corekube_travis.BuildConfigAndCreateStack()
	runTests(config, stackDetails)
	goheat.DeleteStack(config, stackDetails.Stack.Links[0].Href)
}
