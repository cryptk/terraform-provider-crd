package crd

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{},
		ResourcesMap: map[string]*schema.Resource{
			"kubernetes_custom": resourceIstioGateway(),
		},
		DataSourcesMap: map[string]*schema.Resource{},
	}
}
