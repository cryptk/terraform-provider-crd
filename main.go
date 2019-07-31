package main

import (
	"github.com/crainte/terraform-provider-istio/istio"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: crd.Provider})
}
