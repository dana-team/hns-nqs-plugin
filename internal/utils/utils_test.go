package utils_test

import (
	"reflect"
	"testing"
	"time"

	danav1alpha1 "nodeQuotaSync/api/v1alpha1"
	utils "nodeQuotaSync/internal/utils"

	danav1 "github.com/dana-team/hns/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestGetSubnamespaceFromList(t *testing.T) {
	subnamespaceList := danav1.SubnamespaceList{
		Items: []danav1.Subnamespace{
			{},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subnamesapce1",
					Namespace: "test-ns",
				},
			},
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "subnamespace2",
					Namespace: "test-ns",
				},
			},
		},
	}

	t.Run("Existing subnamespace should be returned", func(t *testing.T) {
		subnamespace := utils.GetSubnamespaceFromList("subnamespace2", subnamespaceList)
		if subnamespace == nil {
			t.Error("Expected subnamespace to be returned, but got nil")
		} else if subnamespace.Name != "subnamespace2" {
			t.Errorf("Expected subnamespace with name 'subnamespace2', but got '%s'", subnamespace.Name)
		}
	})

	t.Run("Non-existing subnamespace should return nil", func(t *testing.T) {
		subnamespace := utils.GetSubnamespaceFromList("nonexistent", subnamespaceList)
		if subnamespace != nil {
			t.Errorf("Expected nil, but got subnamespace with name '%s'", subnamespace.Name)
		}
	})
}

func TestHoursPassedSinceDate(t *testing.T) {
	t.Run("Should return correct number of hours passed", func(t *testing.T) {
		timestamp := metav1.Time{
			Time: time.Now().Add(-3 * time.Hour), // Set timestamp 3 hours ago
		}

		hoursPassed := utils.HoursPassedSinceDate(timestamp)
		expectedHoursPassed := int(3)
		if hoursPassed != expectedHoursPassed {
			t.Errorf("Expected %v hours passed, but got %v", expectedHoursPassed, hoursPassed)
		}
	})
}

func TestCalculateGroupReservedResources(t *testing.T) {
	reserved := []danav1alpha1.ReservedResources{
		{
			NodeGroup: "group1",
			Resources: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(2048, resource.BinarySI),
			},
		},
		{
			NodeGroup: "group2",
			Resources: v1.ResourceList{
				v1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
				v1.ResourceMemory: *resource.NewQuantity(4096, resource.BinarySI),
			},
		},
	}

	t.Run("Calculate group reserved resources with existing group", func(t *testing.T) {
		group := "group1"
		expectedResources := v1.ResourceList{
			v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
			v1.ResourceMemory: *resource.NewQuantity(2048, resource.BinarySI),
		}

		calculatedResources := utils.CaculateGroupReservedResources(reserved, group, 24)
		if !reflect.DeepEqual(calculatedResources, expectedResources) {
			t.Errorf("Expected resources %v, but got %v", expectedResources, calculatedResources)
		}
	})

	t.Run("Calculate group reserved resources with non-existing group", func(t *testing.T) {
		group := "nonexistent"
		expectedResources := v1.ResourceList{}

		calculatedResources := utils.CaculateGroupReservedResources(reserved, group, 24)
		if !reflect.DeepEqual(calculatedResources, expectedResources) {
			t.Errorf("Expected resources %v, but got %v", expectedResources, calculatedResources)
		}
	})
}

func TestDeleteExpiredReservedResources(t *testing.T) {
	t.Run("Delete expired reserved resources", func(t *testing.T) {
		config := &danav1alpha1.NodeQuotaConfig{
			Spec: danav1alpha1.NodeQuotaConfigSpec{
				ReservedHoursTolive: 24,
			},
			Status: danav1alpha1.NodeQuotaConfigStatus{
				ReservedResources: []danav1alpha1.ReservedResources{
					{
						Timestamp: metav1.Time{Time: time.Now().Add(-20 * time.Hour)}, // Expired timestamp
						Resources: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(1000, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(2048, resource.BinarySI),
						},
					},
					{
						Timestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)}, // Non-expired timestamp
						Resources: v1.ResourceList{
							v1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
							v1.ResourceMemory: *resource.NewQuantity(4096, resource.BinarySI),
						},
					},
				},
			},
		}

		utils.DeleteExpiredReservedResources(config)

		expectedReservedResources := []danav1alpha1.ReservedResources{
			{
				Timestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				Resources: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(4096, resource.BinarySI),
				},
			},
		}

		if !reflect.DeepEqual(config.Status.ReservedResources, expectedReservedResources) {
			t.Errorf("Expected reserved resources %v, but got %v", expectedReservedResources, config.Status.ReservedResources)
		}
	})
}

func TestMergeTwoResourceList(t *testing.T) {
	resourcesList1 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("1Gi"),
	}

	resourcesList2 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("2Gi"),
	}

	expectedList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("3"),
		v1.ResourceMemory: resource.MustParse("3Gi"),
	}

	result := utils.MergeTwoResourceList(resourcesList1, resourcesList2)

	if result.Cpu() != expectedList.Cpu() || result.Memory() != expectedList.Memory() {
		t.Errorf("MergeTwoResourceList failed, expected: %v, got: %v", expectedList, result)
	}
}

func TestSubstractTwoResourceList(t *testing.T) {
	resourcesList1 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("3"),
		v1.ResourceMemory: resource.MustParse("3Gi"),
	}

	resourcesList2 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("1Gi"),
	}

	expectedList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("2Gi"),
	}

	result := utils.SubstractTwoResourceList(resourcesList1, resourcesList2)

	if !reflect.DeepEqual(expectedList, result) {
		t.Errorf("SubstractTwoResourceList failed, expected: %v, got: %v", expectedList, result)
	}
}

func TestGetPlusAndDebtResourceList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:     resource.MustParse("1"),
		v1.ResourceMemory:  resource.MustParse("-1Gi"),
		v1.ResourceStorage: resource.MustParse("500Mi"),
	}

	expectedPlusResources := v1.ResourceList{
		v1.ResourceCPU:     resource.MustParse("1"),
		v1.ResourceStorage: resource.MustParse("500Mi"),
	}

	expectedDebtResources := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse("-1Gi"),
	}

	plusResources, debtResources := utils.GetPlusAndDebtResourceList(resourcesList)

	if !reflect.DeepEqual(expectedPlusResources, plusResources) {
		t.Errorf("GetPlusAndDebtResourceList (plusResources) failed, expected: %v, got: %v", expectedPlusResources, plusResources)
	}

	if !reflect.DeepEqual(expectedDebtResources, debtResources) {
		t.Errorf("GetPlusAndDebtResourceList (debtResources) failed, expected: %v, got: %v", expectedDebtResources, debtResources)
	}
}

func TestAddResourcesToList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("1Gi"),
	}

	expectedList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("1800M"),
	}

	quantity := resource.MustParse("1")
	name := "cpu"
	utils.AddResourcesToList(&resourcesList, quantity, name)
	utils.AddResourcesToList(&resourcesList, resource.MustParse("800M"), "memory")

	if resourcesList.Cpu().Value() != expectedList.Cpu().Value() {
		t.Errorf("AddResourcesToList failed, expected: %v, got: %v", expectedList, resourcesList)
	}
}

func TestGetResourcesfromList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("1Gi"),
	}

	expectedQuantity := resource.MustParse("1Gi")
	name := "memory"
	quantity := utils.GetResourcesfromList(resourcesList, name)

	if !quantity.Equal(expectedQuantity) {
		t.Errorf("GetResourcesfromList failed, expected: %v, got: %v", expectedQuantity, quantity)
	}
}

func TestMultiplyResourceList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("1Gi"),
	}

	factor := map[string]string{
		"cpu":    "3",
		"memory": "2",
	}

	expectedList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("3"),
		v1.ResourceMemory: resource.MustParse("2Gi"),
	}

	result := utils.MultiplyResourceList(resourcesList, factor)

	if expectedList.Cpu().Value() != result.Cpu().Value() || expectedList.Memory().Value() != result.Memory().Value() {
		t.Errorf("MultiplyResourceList failed, expected: %v, got: %v", expectedList, result)
	}
}
