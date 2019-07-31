package crd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceKubernetesCustom() *schema.Resource {
	return &schema.Resource{
		Create: resourceCRDCreate,
		Read:   resourceCRDRead,
		Exists: resourceCRDExists,
		Update: resourceCRDUpdate,
		Delete: resourceCRDDelete,

		Schema: map[string]*schema.Schema{
			"group": {
				Type:        schema.TypeString,
				Description: "The api group",
				Required:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: "The api version (ex. v1alpha3)",
				Required:    true,
			},
			"resource": {
				Type:        schema.TypeString,
				Description: "The name of the resource",
				Required:    true,
			},
			"file": {
				Type:        schema.TypeString,
				Description: "Path to the yaml file",
				Required:    true,
			},
		},
	}
}

func resourceCRDCreate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCRDRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCRDExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	return true, nil
}

func resourceCRDUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

func resourceCRDDelete(d *schema.ResourceData, meta interface{}) error {
	return nil
}
