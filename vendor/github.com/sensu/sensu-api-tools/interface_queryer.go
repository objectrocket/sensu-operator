package apitools

import (
	"strings"
)

// FindTypesOf finds resources that match a particular interface.
func FindTypesOf[T any]() []T {
	result := []T{}
	IterTypes(func(group, name string, resource any) bool {
		if t, ok := resource.(T); ok && strings.ToLower(name) != name {
			result = append(result, t)
		}
		return true
	})
	return result
}
