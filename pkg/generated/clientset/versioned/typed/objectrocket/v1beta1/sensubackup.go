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

// SensuBackupsGetter has a method to return a SensuBackupInterface.
// A group's client should implement this interface.
type SensuBackupsGetter interface {
	SensuBackups(namespace string) SensuBackupInterface
}

// SensuBackupInterface has methods to work with SensuBackup resources.
type SensuBackupInterface interface {
	Create(*v1beta1.SensuBackup) (*v1beta1.SensuBackup, error)
	Update(*v1beta1.SensuBackup) (*v1beta1.SensuBackup, error)
	UpdateStatus(*v1beta1.SensuBackup) (*v1beta1.SensuBackup, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1beta1.SensuBackup, error)
	List(opts v1.ListOptions) (*v1beta1.SensuBackupList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuBackup, err error)
	SensuBackupExpansion
}

// sensuBackups implements SensuBackupInterface
type sensuBackups struct {
	client rest.Interface
	ns     string
}

// newSensuBackups returns a SensuBackups
func newSensuBackups(c *ObjectrocketV1beta1Client, namespace string) *sensuBackups {
	return &sensuBackups{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the sensuBackup, and returns the corresponding sensuBackup object, and an error if there is any.
func (c *sensuBackups) Get(name string, options v1.GetOptions) (result *v1beta1.SensuBackup, err error) {
	result = &v1beta1.SensuBackup{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sensubackups").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of SensuBackups that match those selectors.
func (c *sensuBackups) List(opts v1.ListOptions) (result *v1beta1.SensuBackupList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1beta1.SensuBackupList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("sensubackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested sensuBackups.
func (c *sensuBackups) Watch(opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("sensubackups").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch()
}

// Create takes the representation of a sensuBackup and creates it.  Returns the server's representation of the sensuBackup, and an error, if there is any.
func (c *sensuBackups) Create(sensuBackup *v1beta1.SensuBackup) (result *v1beta1.SensuBackup, err error) {
	result = &v1beta1.SensuBackup{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("sensubackups").
		Body(sensuBackup).
		Do().
		Into(result)
	return
}

// Update takes the representation of a sensuBackup and updates it. Returns the server's representation of the sensuBackup, and an error, if there is any.
func (c *sensuBackups) Update(sensuBackup *v1beta1.SensuBackup) (result *v1beta1.SensuBackup, err error) {
	result = &v1beta1.SensuBackup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sensubackups").
		Name(sensuBackup.Name).
		Body(sensuBackup).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *sensuBackups) UpdateStatus(sensuBackup *v1beta1.SensuBackup) (result *v1beta1.SensuBackup, err error) {
	result = &v1beta1.SensuBackup{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("sensubackups").
		Name(sensuBackup.Name).
		SubResource("status").
		Body(sensuBackup).
		Do().
		Into(result)
	return
}

// Delete takes name of the sensuBackup and deletes it. Returns an error if one occurs.
func (c *sensuBackups) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sensubackups").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *sensuBackups) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	var timeout time.Duration
	if listOptions.TimeoutSeconds != nil {
		timeout = time.Duration(*listOptions.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("sensubackups").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Timeout(timeout).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched sensuBackup.
func (c *sensuBackups) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuBackup, err error) {
	result = &v1beta1.SensuBackup{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("sensubackups").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
