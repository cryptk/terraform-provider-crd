package crd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	Kschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func resourceCRD() *schema.Resource {
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
			"type": {
				Type:        schema.TypeString,
				Description: "The name of the resource type",
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

	var err error
	var m interface{}
	var un = &unstructured.Unstructured{}

	conn := meta.(dynamic.Interface)

	group := d.Get("group").(string)
	version := d.Get("version").(string)
	resource := d.Get("type").(string)
	file := d.Get("file").(string)

	err = yaml.Unmarshal([]byte(file), &m)
	if err != nil {
		return fmt.Errorf("Failed to create yaml structure: %s", err)
	}
	m = cleanYAML(m)
	un.SetUnstructuredContent(m.(map[string]interface{}))

	dynamicResource := Kschema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Create(un, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to create dynamic resource '%s' because: %s", buildId(un), err)
	}
	log.Printf("[INFO] Submitted new resource: %#v", out)
	d.SetId(buildId(un))

	return resourceCRDRead(d, meta)
}

func resourceCRDRead(d *schema.ResourceData, meta interface{}) error {

	var err error

	conn := meta.(dynamic.Interface)

	group := d.Get("group").(string)
	version := d.Get("version").(string)
	resource := d.Get("type").(string)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	dynamicResource := Kschema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	out, err := conn.Resource(dynamicResource).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		log.Printf("[DEBUG] Received error: %#v", err)
		return fmt.Errorf("Failed to read dynamic resource '%s' because: %s", out, err)
	}
	log.Printf("[INFO] Received resource: %#v", out)

	return nil
}

func resourceCRDExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	var err error

	conn := meta.(dynamic.Interface)

	group := d.Get("group").(string)
	version := d.Get("version").(string)
	resource := d.Get("type").(string)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return false, err
	}

	log.Printf("[INFO] Checking resource %s", name)

	dynamicResource := Kschema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	_, err = conn.Resource(dynamicResource).Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
			return false, nil
		}
		log.Printf("[DEBUG] Received error: %#v", err)
	}
	return true, nil
}

func resourceCRDUpdate(d *schema.ResourceData, meta interface{}) error {

	var err error
	var m interface{}
	var un = &unstructured.Unstructured{}

	conn := meta.(dynamic.Interface)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	group := d.Get("group").(string)
	version := d.Get("version").(string)
	resource := d.Get("type").(string)
	file := d.Get("file").(string)

	err = yaml.Unmarshal([]byte(file), &m)
	if err != nil {
		return fmt.Errorf("Failed to create yaml structure: %s", err)
	}
	m = cleanYAML(m)
	un.SetUnstructuredContent(m.(map[string]interface{}))

	dynamicResource := Kschema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	// dynamic requires resource version to match
	out, err := conn.Resource(dynamicResource).Namespace(namespace).Get(name, metav1.GetOptions{})
	un.SetResourceVersion(out.GetResourceVersion())

	out, err = conn.Resource(dynamicResource).Namespace(namespace).Update(un, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("Failed to update dynamic resource '%s' because: %s", buildId(un), err)
	}
	log.Printf("[INFO] Submitted updated resource: %#v", out)

	return resourceCRDRead(d, meta)
}

func resourceCRDDelete(d *schema.ResourceData, meta interface{}) error {

	var err error

	conn := meta.(dynamic.Interface)

	namespace, name, err := idParts(d.Id())
	if err != nil {
		return err
	}

	log.Printf("[INFO] Deleting resource %#v", name)

	group := d.Get("group").(string)
	version := d.Get("version").(string)
	resource := d.Get("type").(string)

	dynamicResource := Kschema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}

	err = conn.Resource(dynamicResource).Namespace(namespace).Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Failed to delete dynamic resource '%s' because: %s", d.Id(), err)
	}

	log.Printf("[INFO] Resource %s deleted", name)

	d.SetId("")
	return nil
}
