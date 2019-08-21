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

	log.Printf("[DEBUG] In create")
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

		log.Printf("[INFO] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}
		log.Printf("[INFO] APIVersion %#v", apiVersion)

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
			log.Printf("[INFO] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[INFO] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Create(&un, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to create dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[INFO] Submitted new resource: %#v", out)

		// NOTE(crainte): Hackish. This will store the ID as the last item if a list is provided
		d.SetId(out.GetSelfLink())
	}

	return resourceCRDRead(d, meta)
}

func resourceCRDRead(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] In read")
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

		log.Printf("[INFO] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}
		log.Printf("[INFO] APIVersion %#v", apiVersion)

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
			log.Printf("[INFO] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[INFO] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Get(un.GetName(), metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("Failed to read dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[INFO] Received resource: %#v", out)
	}

	return nil
}

func resourceCRDExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	log.Printf("[DEBUG] In exists")
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

		log.Printf("[INFO] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return false, err
		}
		log.Printf("[INFO] APIVersion %#v", apiVersion)

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
			log.Printf("[INFO] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[INFO] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Get(un.GetName(), metav1.GetOptions{})
		if err != nil {
			log.Printf("[DEBUG] err != nil")
			if statusErr, ok := err.(*errors.StatusError); ok && statusErr.ErrStatus.Code == 404 {
				log.Printf("[DEBUG] Error with status: %#v", err)
				return false, nil
			}
			return false, fmt.Errorf("Failed to read dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[INFO] Received resource: %#v", out)
	}
	return true, nil
}

func resourceCRDUpdate(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] In update")
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

		log.Printf("[INFO] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}
		log.Printf("[INFO] APIVersion %#v", apiVersion)

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
			log.Printf("[INFO] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[INFO] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		// requires version to match
		out, err := conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Get(un.GetName(), metav1.GetOptions{})
		un.SetResourceVersion(out.GetResourceVersion())

		out, err = conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Update(&un, metav1.UpdateOptions{})
		if err != nil {
			return fmt.Errorf("Failed to update dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[INFO] Submitted updated resource: %#v", out)
	}

	return resourceCRDRead(d, meta)
}

func resourceCRDDelete(d *schema.ResourceData, meta interface{}) error {

	log.Printf("[DEBUG] In delete")
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

		log.Printf("[INFO] Working with item: %#v", m)
		un.SetUnstructuredContent(m)

		apiVersion, err := Kschema.ParseGroupVersion(m["apiVersion"].(string))
		if err != nil {
			fmt.Errorf("Unable to parse GroupVersion: %s", m["apiVersion"])
			return err
		}
		log.Printf("[INFO] APIVersion %#v", apiVersion)

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
			log.Printf("[INFO] Correcting Group and Version")
			resource.Group = ""
			resource.Version = "v1"
		}

		log.Printf("[INFO] Resource values Group: %s Version: %s Name: %s", resource.Group, resource.Version, resource.Name)

		dynamicResource := Kschema.GroupVersionResource{
			Group:    resource.Group,
			Version:  resource.Version,
			Resource: resource.Name,
		}

		err = conn.Resource(dynamicResource).Namespace(un.GetNamespace()).Delete(un.GetName(), &metav1.DeleteOptions{})
		if err != nil {
			return fmt.Errorf("Failed to delete dynamic resource '%s' because: %s", buildId(&un), err)
		}
		log.Printf("[INFO] Resource %s deleted", un.GetName())
	}

	d.SetId("")
	return nil
}
