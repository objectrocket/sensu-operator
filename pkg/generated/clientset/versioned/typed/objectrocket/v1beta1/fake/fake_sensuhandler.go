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

package fake

import (
	v1beta1 "github.com/objectrocket/sensu-operator/pkg/apis/objectrocket/v1beta1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeSensuHandlers implements SensuHandlerInterface
type FakeSensuHandlers struct {
	Fake *FakeObjectrocketV1beta1
	ns   string
}

var sensuhandlersResource = schema.GroupVersionResource{Group: "objectrocket.com", Version: "v1beta1", Resource: "sensuhandlers"}

var sensuhandlersKind = schema.GroupVersionKind{Group: "objectrocket.com", Version: "v1beta1", Kind: "SensuHandler"}

// Get takes name of the sensuHandler, and returns the corresponding sensuHandler object, and an error if there is any.
func (c *FakeSensuHandlers) Get(name string, options v1.GetOptions) (result *v1beta1.SensuHandler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(sensuhandlersResource, c.ns, name), &v1beta1.SensuHandler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.SensuHandler), err
}

// List takes label and field selectors, and returns the list of SensuHandlers that match those selectors.
func (c *FakeSensuHandlers) List(opts v1.ListOptions) (result *v1beta1.SensuHandlerList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(sensuhandlersResource, sensuhandlersKind, c.ns, opts), &v1beta1.SensuHandlerList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1beta1.SensuHandlerList{}
	for _, item := range obj.(*v1beta1.SensuHandlerList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested sensuHandlers.
func (c *FakeSensuHandlers) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(sensuhandlersResource, c.ns, opts))

}

// Create takes the representation of a sensuHandler and creates it.  Returns the server's representation of the sensuHandler, and an error, if there is any.
func (c *FakeSensuHandlers) Create(sensuHandler *v1beta1.SensuHandler) (result *v1beta1.SensuHandler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(sensuhandlersResource, c.ns, sensuHandler), &v1beta1.SensuHandler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.SensuHandler), err
}

// Update takes the representation of a sensuHandler and updates it. Returns the server's representation of the sensuHandler, and an error, if there is any.
func (c *FakeSensuHandlers) Update(sensuHandler *v1beta1.SensuHandler) (result *v1beta1.SensuHandler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(sensuhandlersResource, c.ns, sensuHandler), &v1beta1.SensuHandler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.SensuHandler), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeSensuHandlers) UpdateStatus(sensuHandler *v1beta1.SensuHandler) (*v1beta1.SensuHandler, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(sensuhandlersResource, "status", c.ns, sensuHandler), &v1beta1.SensuHandler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.SensuHandler), err
}

// Delete takes name of the sensuHandler and deletes it. Returns an error if one occurs.
func (c *FakeSensuHandlers) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(sensuhandlersResource, c.ns, name), &v1beta1.SensuHandler{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeSensuHandlers) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(sensuhandlersResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1beta1.SensuHandlerList{})
	return err
}

// Patch applies the patch and returns the patched sensuHandler.
func (c *FakeSensuHandlers) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1beta1.SensuHandler, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(sensuhandlersResource, c.ns, name, data, subresources...), &v1beta1.SensuHandler{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1beta1.SensuHandler), err
}