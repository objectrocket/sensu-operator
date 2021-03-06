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

// SensuEventFiltersGetter has a method to return a SensuEventFilterInterface.
// A group's client should implement this interface.
type SensuEventFiltersGetter interface {
	SensuEventFilters(namespace string) SensuEventFilterInterface
}

// SensuEventFilterInterface has methods to work with SensuEventFilter resources.
type SensuEventFilterInterface interface {
	Create(*v1beta1.SensuEventFilter) (*v1beta1.SensuEventFilter, error)
	Update(*v1beta1.SensuEventFilter) (*v1beta1.SensuEventFilter, error)
	UpdateStatus(*v1beta1.SensuEventFilter) (*v1beta1.SensuEventFilter, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.SensuEventFilter, error)
	List(opts v1.ListOptions) (*v1beta1.SensuEventFilterList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuEventFilter, err error)
	SensuEventFilterExpansion
}

// sensuEventFilters implements SensuEventFilterInterface
type sensuEventFilters struct {
	client rest.Interface
	ns     string
}

// newSensuEventFilters returns a SensuEventFilters
func newSensuEventFilters(c *ObjectrocketV1beta1Client, namespace string) *sensuEventFilters {
	return &sensuEventFilters{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sensuEventFilter, and returns the corresponding sensuEventFilter object, and an error if there is any.
func (c *sensuEventFilters) Get(name string, options v1.GetOptions) (result *v1beta1.SensuEventFilter, err error) {
	result = &v1beta1.SensuEventFilter{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sensueventfilters").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SensuEventFilters that match those selectors.
func (c *sensuEventFilters) List(opts v1.ListOptions) (result *v1beta1.SensuEventFilterList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.SensuEventFilterList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sensueventfilters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sensuEventFilters.
func (c *sensuEventFilters) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sensueventfilters").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a sensuEventFilter and creates it.  Returns the server's representation of the sensuEventFilter, and an error, if there is any.
func (c *sensuEventFilters) Create(sensuEventFilter *v1beta1.SensuEventFilter) (result *v1beta1.SensuEventFilter, err error) {
	result = &v1beta1.SensuEventFilter{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sensueventfilters").
		Body(sensuEventFilter).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sensuEventFilter and updates it. Returns the server's representation of the sensuEventFilter, and an error, if there is any.
func (c *sensuEventFilters) Update(sensuEventFilter *v1beta1.SensuEventFilter) (result *v1beta1.SensuEventFilter, err error) {
	result = &v1beta1.SensuEventFilter{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sensueventfilters").
		Name(sensuEventFilter.Name).
		Body(sensuEventFilter).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *sensuEventFilters) UpdateStatus(sensuEventFilter *v1beta1.SensuEventFilter) (result *v1beta1.SensuEventFilter, err error) {
	result = &v1beta1.SensuEventFilter{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sensueventfilters").
		Name(sensuEventFilter.Name).
		SubResource("status").
		Body(sensuEventFilter).
		Do().
		Into(result)
	return
}

// Delete takes name of the sensuEventFilter and deletes it. Returns an error if one occurs.
func (c *sensuEventFilters) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sensueventfilters").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sensuEventFilters) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sensueventfilters").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sensuEventFilter.
func (c *sensuEventFilters) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuEventFilter, err error) {
	result = &v1beta1.SensuEventFilter{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sensueventfilters").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
