package main

import (
	"github.com/crainte/terraform-provider-crd/crd"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: crd.Provider})
}
