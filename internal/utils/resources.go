package utils

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/strings/slices"
)

// FilterUncontrolledResources filters the given resources list based on the controlled resources.
// It returns a new resource list that contains only the resources specified in the controlled resources list.
func FillterUncontrolledResources(resourcesList v1.ResourceList, controlledResoures []string) v1.ResourceList {
	fillteredList := v1.ResourceList{}
	for resourceName, quantity := range resourcesList {
		if slices.Contains(controlledResoures, resourceName.String()) {
			AddResourcesToList(&fillteredList, quantity, resourceName.String())
		}
	}
	return fillteredList
}

// AddResourcesToList adds the given quantity of a resource with the specified name to the resource list.
// If the resource with the same name already exists in the list, it adds the quantity to the existing resource.
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

// GetResourcesFromList retrieves the quantity of a resource with the specified name from the resource list.
// It returns the quantity if found, otherwise it returns a zero quantity.
func GetResourcesfromList(resourcesList v1.ResourceList, name string) resource.Quantity {
	for resourceName, resourceQuantity := range resourcesList {
		if name == string(resourceName) {
			return resourceQuantity
		}
	}
	return resource.Quantity{}
}

// MergeTwoResourceList merges two resource lists into a single resource list.
// It combines the quantities of the same resources from both lists.
func MergeTwoResourceList(resourcelist v1.ResourceList, resourcelist2 v1.ResourceList) v1.ResourceList {
	result := make(v1.ResourceList)

	// Merge quantities from resourcelist
	for resourceName, resourceQuantity := range resourcelist {
		result[resourceName] = resourceQuantity.DeepCopy()
	}

	// Merge quantities from resourcelist2
	for resourceName, resourceQuantity := range resourcelist2 {
		AddResourcesToList(&result, resourceQuantity, string(resourceName))
	}

	return result
}

// SubtractTwoResourceList subtracts the quantities of resources in resourcelist2 from resourcelist.
// It returns a new resource list with the subtracted quantities.
func SubstractTwoResourceList(resourcelist v1.ResourceList, resourcelist2 v1.ResourceList) v1.ResourceList {
	newResourceList := resourcelist.DeepCopy()
	for resourceName, resourceQuantity := range newResourceList {
		resourceQuantity.Sub(GetResourcesfromList(resourcelist2, string(resourceName)))
	}
	return newResourceList

}

// GetPlusAndDebtResourceList categorizes the resources in the given resource list into two separate lists:
// plusResources (resources with positive quantities) and debtResources (resources with negative quantities).
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

// MultiplyResourceList multiplies the values of resources in the given resource list by the corresponding factors.
// It returns a new resource list with the multiplied values.
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
