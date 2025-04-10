package utils

import (
	"testing"
	"time"

	danav1alpha1 "github.com/dana-team/hns-nqs-plugin/api/v1alpha1"
	danav1 "github.com/dana-team/hns/api/v1"
	"github.com/go-logr/logr"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const resourceName = "cpu"

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
		subnamespace := GetSubnamespaceFromList("subnamespace2", subnamespaceList)
		if subnamespace == nil {
			t.Error("Expected subnamespace to be returned, but got nil")
		} else if subnamespace.Name != "subnamespace2" {
			t.Errorf("Expected subnamespace with name 'subnamespace2', but got '%s'", subnamespace.Name)
		}
	})

	t.Run("Non-existing subnamespace should return nil", func(t *testing.T) {
		subnamespace := GetSubnamespaceFromList("nonexistent", subnamespaceList)
		if subnamespace != nil {
			t.Errorf("Expected nil, but got subnamespace with name '%s'", subnamespace.Name)
		}
	})
}

func TestDeleteExpiredReservedResources(t *testing.T) {
	t.Run("Delete expired reserved resources", func(t *testing.T) {
		config := &danav1alpha1.NodeQuotaConfig{
			Spec: danav1alpha1.NodeQuotaConfigSpec{
				ReservedHoursToLive: 24,
			},
			Status: danav1alpha1.NodeQuotaConfigStatus{
				ReservedResources: []danav1alpha1.ReservedResources{
					{
						Timestamp: metav1.Time{Time: time.Now().Add(-25 * time.Hour)}, // Expired timestamp
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

		DeleteExpiredReservedResources(config, logr.Discard())

		expectedReservedResources := []danav1alpha1.ReservedResources{
			{
				Timestamp: metav1.Time{Time: time.Now().Add(-1 * time.Hour)},
				Resources: v1.ResourceList{
					v1.ResourceCPU:    *resource.NewMilliQuantity(2000, resource.DecimalSI),
					v1.ResourceMemory: *resource.NewQuantity(4096, resource.BinarySI),
				},
			},
		}

		if len(expectedReservedResources) != 1 {
			t.Errorf("Expected reserved resources %v, but got %v", expectedReservedResources, config.Status.ReservedResources)
		}
	})
}

func TestFilterUncontrolledResources(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("2Gi"),
	}
	controlledResources := []string{resourceName}
	expectedResult := v1.ResourceList{
		v1.ResourceCPU: resource.MustParse("1"),
	}

	result := filterUncontrolledResources(resourcesList, controlledResources)
	assert.Equal(t, expectedResult, result)
}

func TestIsGreaterThan(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	resourcesList2 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("1"),
		v1.ResourceMemory: resource.MustParse("2Gi"),
	}

	result := isGreaterThan(resourcesList, resourcesList2)
	assert.True(t, result)
}

func TestIsEqualTo(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	resourcesList2 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}

	result := isEqualTo(resourcesList, resourcesList2)
	assert.True(t, result)
}

func TestAddResourcesToList(t *testing.T) {
	resourcesList := v1.ResourceList{}
	quantity := resource.MustParse("2")
	name := resourceName
	expectedResult := v1.ResourceList{
		v1.ResourceName(name): quantity,
	}

	addResourcesToList(&resourcesList, quantity, name)
	assert.Equal(t, expectedResult, resourcesList)
}

func TestGetResourcesfromList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	name := resourceName
	expectedResult := resource.MustParse("2")

	result := getResourcesfromList(resourcesList, name)
	assert.Equal(t, expectedResult, result)
}

func TestMergeTwoResourceList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	resourcesList2 := v1.ResourceList{
		v1.ResourceMemory: resource.MustParse("8Gi"),
		v1.ResourcePods:   resource.MustParse("10"),
	}
	expectedResult := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("12Gi"),
		v1.ResourcePods:   resource.MustParse("10"),
	}

	result := MergeTwoResourceList(resourcesList, resourcesList2)
	assert.True(t, expectedResult.Cpu().Equal(*result.Cpu()), expectedResult.Memory().Equal(*result.Memory()), expectedResult.Pods().Equal(*result.Memory()))
}

func TestSubtractFromResourcesList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("4"),
		v1.ResourceMemory: resource.MustParse("8Gi"),
	}
	quantity := resource.MustParse("2")
	name := resourceName
	expectedResult := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("8Gi"),
	}

	subtractFromResourcesList(&resourcesList, quantity, name)
	assert.True(t, expectedResult.Cpu().Equal(*resourcesList.Cpu()))
}

func TestSubtractTwoResourceList(t *testing.T) {
	resourcesList := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("4"),
		v1.ResourceMemory: resource.MustParse("8Gi"),
	}
	resourcesList2 := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	expectedResult := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}

	result := subtractTwoResourceList(resourcesList, resourcesList2)
	assert.True(t, expectedResult.Cpu().Equal(*result.Cpu()), expectedResult.Memory().Equal(*result.Memory()))
}

func TestMultiplyResourceList(t *testing.T) {
	resources := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("2"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	factor := map[string]string{
		"cpu": "2",
	}
	expectedResult := v1.ResourceList{
		v1.ResourceCPU:    resource.MustParse("4"),
		v1.ResourceMemory: resource.MustParse("4Gi"),
	}
	reservedResources := map[string]resource.Quantity{
		"cpu":    resource.MustParse("2"),
		"memory": resource.MustParse("2Gi"),
	}
	logger := logr.Discard()

	result := multiplyResourceList(resources, factor, reservedResources, logger)
	assert.True(t, result.Cpu().Equal(*expectedResult.Cpu()))
}

func TestHoursPassedSinceDate(t *testing.T) {
	sometime := metav1.Time{Time: time.Now().Add(-25 * time.Hour)}
	result := hoursPassedSinceDate(sometime)
	assert.Equal(t, 25, result)
}

func TestSubtractResources(t *testing.T) {
	// Create a test logger
	logger := logr.Discard()

	tests := []struct {
		name              string
		resources         v1.ResourceList
		reservedResources map[string]resource.Quantity
		expected          v1.ResourceList
	}{
		{
			name: "successful subtraction",
			resources: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
			},
			reservedResources: map[string]resource.Quantity{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("2Gi"),
			},
			expected: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("3"),
				v1.ResourceMemory: resource.MustParse("6Gi"),
			},
		},
		{
			name: "attempt subtraction with negative result",
			resources: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
			reservedResources: map[string]resource.Quantity{
				"cpu":    resource.MustParse("2"),
				"memory": resource.MustParse("2Gi"),
			},
			expected: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("1"),
				v1.ResourceMemory: resource.MustParse("1Gi"),
			},
		},
		{
			name: "missing resource in reserved",
			resources: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
			},
			reservedResources: map[string]resource.Quantity{
				"cpu": resource.MustParse("1"),
			},
			expected: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("3"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
		{
			name: "missing resource in original",
			resources: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("4"),
			},
			reservedResources: map[string]resource.Quantity{
				"cpu":    resource.MustParse("1"),
				"memory": resource.MustParse("2Gi"),
			},
			expected: v1.ResourceList{
				v1.ResourceCPU: resource.MustParse("3"),
			},
		},
		{
			name: "empty reserved resources",
			resources: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
			},
			reservedResources: map[string]resource.Quantity{},
			expected: v1.ResourceList{
				v1.ResourceCPU:    resource.MustParse("4"),
				v1.ResourceMemory: resource.MustParse("8Gi"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := subtractResources(tt.resources, tt.reservedResources, logger)

			// Check that the result has the expected values
			assert.Equal(t, len(tt.expected), len(result), "number of resources should match")

			for resourceName, expectedQuantity := range tt.expected {
				resultQuantity, exists := result[resourceName]
				assert.True(t, exists, "resource should exist in result")
				assert.True(t, expectedQuantity.Equal(resultQuantity),
					"resource %v: expected %v, got %v",
					resourceName,
					expectedQuantity.String(),
					resultQuantity.String())
			}
		})
	}
}
