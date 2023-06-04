package utils

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func AddResourcesToList(resourcesList *v1.ResourceList, quantity resource.Quantity, name string) {
	for resourceName, resourceQuantity := range *resourcesList {
		if name == string(resourceName) {
			resourceQuantity.Add(quantity)
		}
	}
}

func GetResourcesfromList(resourcesList v1.ResourceList, name string) resource.Quantity {
	for resourceName, resourceQuantity := range resourcesList {
		if name == string(resourceName) {
			return resourceQuantity
		}
	}
	return resource.Quantity{}
}

func SubstractTwoResourceList(resourcelist v1.ResourceList, resourcelist2 v1.ResourceList) v1.ResourceList {
	newResourceList := resourcelist.DeepCopy()
	for resourceName, resourceQuantity := range newResourceList {
		resourceQuantity.Sub(GetResourcesfromList(resourcelist2, string(resourceName)))
	}
	return newResourceList

}

func MultiplyResourceList(resources v1.ResourceList, factor map[string]float64) v1.ResourceList {
	result := make(v1.ResourceList)

	for name, value := range resources {
		if factor[name.String()] == 0 {
			continue
		}
		newValue := new(resource.Quantity)
		newValue.Set(value.Value())

		// Multiply the value by the factor
		newValueInt64 := float64(newValue.Value()) * factor[name.String()]
		newValue.Set(int64(newValueInt64))

		result[name] = *newValue
	}

	return result
}
