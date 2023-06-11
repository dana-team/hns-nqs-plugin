package utils

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func AddResourcesToList(resourcesList *v1.ResourceList, quantity resource.Quantity, name string) {
	for resourceName, resourceQuantity := range *resourcesList {
		if name == string(resourceName) {
			resourceQuantity.Add(quantity)
			(*resourcesList)[resourceName] = resourceQuantity
			return
		}
	}
	(*resourcesList)[v1.ResourceName(name)] = quantity
}

func GetResourcesfromList(resourcesList v1.ResourceList, name string) resource.Quantity {
	for resourceName, resourceQuantity := range resourcesList {
		if name == string(resourceName) {
			return resourceQuantity
		}
	}
	return resource.Quantity{}
}

func MergeTwoResourceList(resourcelist v1.ResourceList, resourcelist2 v1.ResourceList) v1.ResourceList {
	result := make(v1.ResourceList)

	// Merge quantities from resourcelist
	for resourceName, resourceQuantity := range resourcelist {
		result[resourceName] = resourceQuantity.DeepCopy()
	}

	// Merge quantities from resourcelist2
	for resourceName, resourceQuantity := range resourcelist2 {
		if existingQuantity, found := result[resourceName]; found {
			existingQuantity.Add(resourceQuantity)
		} else {
			result[resourceName] = resourceQuantity.DeepCopy()
		}
	}

	return result
}

func SubstractTwoResourceList(resourcelist v1.ResourceList, resourcelist2 v1.ResourceList) v1.ResourceList {
	newResourceList := resourcelist.DeepCopy()
	for resourceName, resourceQuantity := range newResourceList {
		resourceQuantity.Sub(GetResourcesfromList(resourcelist2, string(resourceName)))
	}
	return newResourceList

}

func GetPlusAndDebtResourceList(resourcelist v1.ResourceList) (v1.ResourceList, v1.ResourceList) {
	debtResources := v1.ResourceList{}
	plusResources := v1.ResourceList{}
	for name, resource := range resourcelist {
		if resource.Sign() > 0 {
			AddResourcesToList(&plusResources, resource, name.String())
		}
		if resource.Sign() < 0 {
			AddResourcesToList(&debtResources, resource, name.String())
		}
	}
	return plusResources, debtResources
}

func MultiplyResourceList(resources v1.ResourceList, factor map[string]string) v1.ResourceList {
	result := make(v1.ResourceList)

	for name, value := range resources {
		if factor[name.String()] == "" {
			result[name] = value
			continue
		}
		newValue := new(resource.Quantity)
		newValue.Set(value.Value())

		// Multiply the value by the factor
		factorFloat, _ := strconv.ParseFloat(factor[name.String()], 64)
		newValueInt64 := float64(newValue.Value()) * factorFloat
		newValue.Set(int64(newValueInt64))

		result[name] = *newValue
	}

	return result
}
