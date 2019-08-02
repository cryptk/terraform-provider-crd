package crd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	Kschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
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
				ForceNew:    true,
			},
			"version": {
				Type:        schema.TypeString,
				Description: "The api version (ex. v1alpha3)",
				Required:    true,
				ForceNew:    true,
			},
			"resource": {
				Type:        schema.TypeString,
				Description: "The name of the resource",
				Required:    true,
				ForceNew:    true,
			},
			"namespace": {
				Type:        schema.TypeString,
				Description: "The namespace of the resource",
				Required:    true,
				ForceNew:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Name of the resource",
				Required:    true,
				ForceNew:    true,
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
	conn := meta.(dynamic.Interface)

	group := d.Get("group")
	version := d.Get("version")
	resource := d.Get("resource")
	namespace := d.Get("namespace")
	name := d.Get("name")

	dynamicResource := Kschema.GroupVersionResource{
		Group:    group.(string),
		Version:  version.(string),
		Resource: resource.(string),
	}

	dyn, err := conn.Resource(dynamicResource).Namespace(namespace.(string)).Get(name.(string), metav1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return fmt.Errorf("Failed to read dynamic resource '%s' because: %s", dyn, err)
	}
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
