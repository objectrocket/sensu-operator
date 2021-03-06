/*
Copyright 2019 The sensu-operator Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by client-gen. DO NOT EDIT.

package v1beta1

import (
	"time"

	v1beta1 "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
	scheme "github.com/objectrocket/sensu-operator/pkg/generated/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// SensuClustersGetter has a method to return a SensuClusterInterface.
// A group's client should implement this interface.
type SensuClustersGetter interface {
	SensuClusters(namespace string) SensuClusterInterface
}

// SensuClusterInterface has methods to work with SensuCluster resources.
type SensuClusterInterface interface {
	Create(*v1beta1.SensuCluster) (*v1beta1.SensuCluster, error)
	Update(*v1beta1.SensuCluster) (*v1beta1.SensuCluster, error)
	UpdateStatus(*v1beta1.SensuCluster) (*v1beta1.SensuCluster, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.SensuCluster, error)
	List(opts v1.ListOptions) (*v1beta1.SensuClusterList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuCluster, err error)
	SensuClusterExpansion
}

// sensuClusters implements SensuClusterInterface
type sensuClusters struct {
	client rest.Interface
	ns     string
}

// newSensuClusters returns a SensuClusters
func newSensuClusters(c *ObjectrocketV1beta1Client, namespace string) *sensuClusters {
	return &sensuClusters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sensuCluster, and returns the corresponding sensuCluster object, and an error if there is any.
func (c *sensuClusters) Get(name string, options v1.GetOptions) (result *v1beta1.SensuCluster, err error) {
	result = &v1beta1.SensuCluster{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sensuclusters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SensuClusters that match those selectors.
func (c *sensuClusters) List(opts v1.ListOptions) (result *v1beta1.SensuClusterList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.SensuClusterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sensuclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sensuClusters.
func (c *sensuClusters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sensuclusters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a sensuCluster and creates it.  Returns the server's representation of the sensuCluster, and an error, if there is any.
func (c *sensuClusters) Create(sensuCluster *v1beta1.SensuCluster) (result *v1beta1.SensuCluster, err error) {
	result = &v1beta1.SensuCluster{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sensuclusters").
		Body(sensuCluster).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sensuCluster and updates it. Returns the server's representation of the sensuCluster, and an error, if there is any.
func (c *sensuClusters) Update(sensuCluster *v1beta1.SensuCluster) (result *v1beta1.SensuCluster, err error) {
	result = &v1beta1.SensuCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sensuclusters").
		Name(sensuCluster.Name).
		Body(sensuCluster).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *sensuClusters) UpdateStatus(sensuCluster *v1beta1.SensuCluster) (result *v1beta1.SensuCluster, err error) {
	result = &v1beta1.SensuCluster{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sensuclusters").
		Name(sensuCluster.Name).
		SubResource("status").
		Body(sensuCluster).
		Do().
		Into(result)
	return
}

// Delete takes name of the sensuCluster and deletes it. Returns an error if one occurs.
func (c *sensuClusters) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sensuclusters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sensuClusters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sensuclusters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sensuCluster.
func (c *sensuClusters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuCluster, err error) {
	result = &v1beta1.SensuCluster{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sensuclusters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
