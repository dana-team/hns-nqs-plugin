package utils

import (
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/strings/slices"
)

// filterUncontrolledResources filters the given resources list based on the controlled resources.
// It returns a new resource list that contains only the resources specified in the controlled resources list.
func filterUncontrolledResources(resourcesList v1.ResourceList, controlledResources []string) v1.ResourceList {
	filteredList := v1.ResourceList{}
	for resourceName, quantity := range resourcesList {
		if slices.Contains(controlledResources, resourceName.String()) {
			addResourcesToList(&filteredList, quantity, resourceName.String())
		}
	}
	return filteredList
}

// isGreaterThan checks if the quantities in resourcesList are greater than or equal to the corresponding quantities in resourcesList2.
func isGreaterThan(resourcesList v1.ResourceList, resourcesList2 v1.ResourceList) bool {
	for resourceName, resourceQuantity := range resourcesList {
		sign := resourceQuantity.Cmp(resourcesList2[resourceName])
		if sign == -1 || sign == 0 {
			return false
		}
	}
	return true
}

// isEqualTo checks if the quantities in resourcesList are equal to the corresponding quantities in resourcesList2.
func isEqualTo(resourcesList v1.ResourceList, resourcesList2 v1.ResourceList) bool {
	for resourceName, resourceQuantity := range resourcesList {
		sign := resourceQuantity.Cmp(resourcesList2[resourceName])
		if sign != 0 {
			return false
		}
	}
	return true
}

// addResourcesToList adds the given quantity of a resource with the specified name to the resource list.
// If the resource with the same name already exists in the list, it adds the quantity to the existing resource.
func addResourcesToList(resourcesList *v1.ResourceList, quantity resource.Quantity, name string) {
	for resourceName, resourceQuantity := range *resourcesList {
		if name == string(resourceName) {
			resourceQuantity.Add(quantity)
			(*resourcesList)[resourceName] = resourceQuantity
			return
		}
	}
	(*resourcesList)[v1.ResourceName(name)] = quantity
}

// getResourcesFromList retrieves the quantity of a resource with the specified name from the resource list.
// It returns the quantity if found, otherwise it returns a zero quantity.
func getResourcesfromList(resourcesList v1.ResourceList, name string) resource.Quantity {
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
		addResourcesToList(&result, resourceQuantity, string(resourceName))
	}

	return result
}

// subtractFromResourcesList subtracts the given quantity from the resource quantity identified by the given name in the resourcesList.
func subtractFromResourcesList(resourcesList *v1.ResourceList, quantity resource.Quantity, name string) {
	for resourceName, resourceQuantity := range *resourcesList {
		if name == string(resourceName) {
			resourceQuantity.Sub(quantity)
			(*resourcesList)[resourceName] = resourceQuantity
			return
		}
	}
	(*resourcesList)[v1.ResourceName(name)] = quantity
}

// subtractTwoResourceList subtracts the quantities of resources in resourcelist2 from resourcelist.
// It returns a new resource list with the subtracted quantities.
func subtractTwoResourceList(resourcesList v1.ResourceList, resourcelist2 v1.ResourceList) v1.ResourceList {
	result := make(v1.ResourceList)

	for resourceName, resourceQuantity := range resourcesList {
		result[resourceName] = resourceQuantity.DeepCopy()
	}

	for resourceName, resourceQuantity := range resourcelist2 {
		subtractFromResourcesList(&result, resourceQuantity, string(resourceName))
	}

	return result
}

func patchResourcesToList(resourcesList v1.ResourceList, resourcesToPatch v1.ResourceList) v1.ResourceList {
	for resourceName, resourceQuantity := range resourcesToPatch {
		resourcesList[resourceName] = resourceQuantity
	}
	return resourcesList
}

// multiplyResourceList multiplies the values of resources in the given resource list by the corresponding factors.
// It returns a new resource list with the multiplied values.
func multiplyResourceList(resources v1.ResourceList, factor map[string]string) v1.ResourceList {
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
