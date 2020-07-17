package crd

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	Kschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	//"k8s.io/client-go/dynamic"
)

func resourceCRD() *schema.Resource {
	return &schema.Resource{
		Create: resourceCRDCreate,
		Read:   resourceCRDRead,
		Exists: resourceCRDExists,
		Update: resourceCRDUpdate,
		Delete: resourceCRDDelete,

		Schema: map[string]*schema.Schema{
			"yaml": {
				Type:        schema.TypeString,
				Description: "Path to the yaml file",
				Required:    true,
			},
		},
	}
}

func resourceCRDCreate(d *schema.ResourceData, meta interface{}) error {

	var un = unstructured.Unstructured{}
	var err error

	clientset := meta.(*KubeClientSet).Main
	conn := meta.(*KubeClientSet).Dynamic

	content := d.Get("yaml").(string)

	reader := strings.NewReader(content)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 10)

	for {
		m := make(map[string]interface{})

		err = decoder.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}

		err = discovery.ServerSupportsVersion(clientset.Discovery(), apiVersion)
		if err != nil {
			fmt.Errorf("Server does not support ApiVersion: %v", apiVersion)
			return err
		}
		resources, err := clientset.Discovery().ServerResources()
		if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return err
		}

		resource, exists := ResourceExists(resources, un)
		if !exists {
			return fmt.Errorf("resource provided in yaml isn't valid for cluster. Ensure APIVersion and Kind are valid")
		}

		if resource.Group == "v1" && resource.Version == "" {
			log.Printf("---[ CRD ]------------------------\n")
			log.Printf("Correcting Group and Version")
			log.Printf("---[ CRD ]------------------------\n")
			resource.Group = ""
			resource.Version = "v1"
		} else {
			resource.Group = apiVersion.Group
			resource.Version = apiVersion.Version
		}

		log.Printf("[CRD] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		namespace := un.GetNamespace()
		if namespace == "" {
			return fmt.Errorf("Failed to find Namespace in yaml")
		}

		out, err := conn.Resource(dynamicResource).Namespace(namespace).Create(&un, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to create dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[CRD] Submitted new resource: %#v", out)

		// NOTE(crainte): Hackish. This will store the ID as the last item if a list is provided
		d.SetId(out.GetSelfLink())
	}

	return resourceCRDRead(d, meta)
}

func resourceCRDRead(d *schema.ResourceData, meta interface{}) error {

	var un = unstructured.Unstructured{}
	var err error

	clientset := meta.(*KubeClientSet).Main
	conn := meta.(*KubeClientSet).Dynamic

	content := d.Get("yaml").(string)

	reader := strings.NewReader(content)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 10)

	for {
		m := make(map[string]interface{})

		err = decoder.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		log.Printf("[CRD] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}
		log.Printf("[CRD] APIVersion %#v", apiVersion)

		err = discovery.ServerSupportsVersion(clientset.Discovery(), apiVersion)
		if err != nil {
			fmt.Errorf("Server does not support ApiVersion: %v", apiVersion)
			return err
		}
		//resources, _ := conn.Discovery().ServerResourcesForGroupVersion(apiVersion.String())
		resources, err := clientset.Discovery().ServerResources()
		if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return err
		}

		resource, exists := ResourceExists(resources, un)
		if !exists {
			return fmt.Errorf("resource provided in yaml isn't valid for cluster. Ensure APIVersion and Kind are valid")
		}

		if resource.Group == "v1" && resource.Version == "" {
			log.Printf("[CRD] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[CRD] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Get(un.GetName(), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to read dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[CRD] Received resource: %#v", out)
	}

	return nil
}

func resourceCRDExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	var un = unstructured.Unstructured{}
	var err error

	clientset := meta.(*KubeClientSet).Main
	conn := meta.(*KubeClientSet).Dynamic

	content := d.Get("yaml").(string)

	reader := strings.NewReader(content)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 10)

	for {
		m := make(map[string]interface{})

		err = decoder.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			return false, err
		}

		log.Printf("[CRD] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return false, err
		}
		log.Printf("[CRD] APIVersion %#v", apiVersion)

		err = discovery.ServerSupportsVersion(clientset.Discovery(), apiVersion)
		if err != nil {
			fmt.Errorf("Server does not support ApiVersion: %v", apiVersion)
			return false, err
		}
		//resources, _ := conn.Discovery().ServerResourcesForGroupVersion(apiVersion.String())
		resources, err := clientset.Discovery().ServerResources()
		if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return false, err
		}

		resource, exists := ResourceExists(resources, un)
		if !exists {
			return false, fmt.Errorf("resource provided in yaml isn't valid for cluster. Ensure APIVersion and Kind are valid")
		}

		if resource.Group == "v1" && resource.Version == "" {
			log.Printf("[CRD] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[CRD] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Get(un.GetName(), metav1.GetOptions{})
		if err != nil {
			log.Printf("[CRD] err != nil")
			if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
				log.Printf("[CRD] Error with status: %#v", err)
				return false, nil
			}
			return false, fmt.Errorf("Failed to read dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[CRD] Received resource: %#v", out)
	}
	return true, nil
}

func resourceCRDUpdate(d *schema.ResourceData, meta interface{}) error {

	var o_unstruct = unstructured.Unstructured{}
	var c_unstruct = unstructured.Unstructured{}
	var err error

	clientset := meta.(*KubeClientSet).Main
	conn := meta.(*KubeClientSet).Dynamic

	if !d.HasChange("yaml") {
		log.Printf("---[ CRD ]------------------------\n")
		log.Printf("\tThere is no change in the yaml")
		log.Printf("---[ CRD ]------------------------\n")
		return nil
	}

	o, c := d.GetChange("yaml")

	o_reader := strings.NewReader(o.(string))
	c_reader := strings.NewReader(c.(string))

	o_decoder := yaml.NewYAMLOrJSONDecoder(o_reader, 10)
	c_decoder := yaml.NewYAMLOrJSONDecoder(c_reader, 10)

	for {
		original := make(map[string]interface{})
		change := make(map[string]interface{})

		err = o_decoder.Decode(&original)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		err = c_decoder.Decode(&change)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		o_unstruct.SetUnstructuredContent(original)
		c_unstruct.SetUnstructuredContent(change)

		apiVersion, err := Kschema.ParseGroupVersion(change["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", change["apiVersion"])
			return err
		}

		err = discovery.ServerSupportsVersion(clientset.Discovery(), apiVersion)
		if err != nil {
			fmt.Errorf("Server does not support ApiVersion: %v", apiVersion)
			return err
		}

		resources, err := clientset.Discovery().ServerResources()
		if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return err
		}

		o_resource, _ := ResourceExists(resources, o_unstruct)
		c_resource, exists := ResourceExists(resources, c_unstruct)
		if !exists {
			return fmt.Errorf("Resource provided in yaml isn't valid for cluster. Ensure APIVersion and Kind are valid")
		}

		if c_resource.Group == "v1" && c_resource.Version == "" {
			log.Printf("---[ CRD ]------------------------\n")
			log.Printf("\tCorrecting Group and Version")
			log.Printf("---[ CRD ]------------------------\n")
			c_resource.Group = ""
			c_resource.Version = "v1"
		}

		log.Printf("---[ CRD ]------------------------\n")
		log.Printf("\tCreating Dynamic Resource\n\tGroup: %s\n\tVersion: %s\n\tName: %s\n", o_resource.Group, o_resource.Version, o_resource.Name)
		log.Printf("\tCreating Dynamic Resource\n\tGroup: %s\n\tVersion: %s\n\tName: %s\n", c_resource.Group, c_resource.Version, c_resource.Name)
		log.Printf("---[ CRD ]------------------------\n")

		o_dynamicResource := Kschema.GroupVersionResource{
			Group:    o_resource.Group,
			Version:  o_resource.Version,
			Resource: o_resource.Name,
		}

		c_dynamicResource := Kschema.GroupVersionResource{
			Group:    c_resource.Group,
			Version:  c_resource.Version,
			Resource: c_resource.Name,
		}

		// delete the original resource
		err = conn.Resource(o_dynamicResource).Namespace(o_unstruct.GetNamespace()).Delete(o_unstruct.GetName(), &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Failed to remove the original object: %s/%s", o_unstruct.GetNamespace(), o_unstruct.GetName())
		}

		// create the new resource
		c_object, err := conn.Resource(c_dynamicResource).Namespace(c_unstruct.GetNamespace()).Create(&c_unstruct, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to create dynamic resource '%s' because: %s", buildId(&c_unstruct), err)
		}
		log.Printf("---[ CRD ]------------------------\n")
		log.Printf("\tSubmitted updated resource:\n\t%+v\n", c_object)
		log.Printf("---[ CRD ]------------------------\n")
	}

	return resourceCRDRead(d, meta)
}

func resourceCRDDelete(d *schema.ResourceData, meta interface{}) error {

	var un = unstructured.Unstructured{}
	var err error

	clientset := meta.(*KubeClientSet).Main
	conn := meta.(*KubeClientSet).Dynamic

	content := d.Get("yaml").(string)

	reader := strings.NewReader(content)
	decoder := yaml.NewYAMLOrJSONDecoder(reader, 10)

	for {
		m := make(map[string]interface{})

		err = decoder.Decode(&m)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		log.Printf("[CRD] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}
		log.Printf("[CRD] APIVersion %#v", apiVersion)

		err = discovery.ServerSupportsVersion(clientset.Discovery(), apiVersion)
		if err != nil {
			fmt.Errorf("Server does not support ApiVersion: %v", apiVersion)
			return err
		}
		//resources, _ := conn.Discovery().ServerResourcesForGroupVersion(apiVersion.String())
		resources, err := clientset.Discovery().ServerResources()
		if err != nil && !discovery.IsGroupDiscoveryFailedError(err) {
			return err
		}

		resource, exists := ResourceExists(resources, un)
		if !exists {
			return fmt.Errorf("resource provided in yaml isn't valid for cluster. Ensure APIVersion and Kind are valid")
		}

		if resource.Group == "v1" && resource.Version == "" {
			log.Printf("[CRD] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[CRD] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		err = conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Delete(un.GetName(), &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Failed to delete dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[CRD] Resource %s deleted", un.GetName())
	}

	d.SetId("")
	return nil
}
