package executor

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.ibm.com/turbonomic/turbo-go-sdk/pkg/proto"
	k8sapi "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createPod() *k8sapi.Pod {
	pod := &k8sapi.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},

		ObjectMeta: metav1.ObjectMeta{
			Name: "my-pod-1",
			UID:  "my-pod-1-UID",
		},

		Spec: k8sapi.PodSpec{},
	}

	resourceSpec := k8sapi.ResourceRequirements{
		Limits:   make(k8sapi.ResourceList),
		Requests: make(k8sapi.ResourceList),
	}

	container := k8sapi.Container{
		Resources: resourceSpec,
	}

	containers := []k8sapi.Container{container}
	pod.Spec.Containers = containers

	return pod
}

// Test Case: resize Memory Capacity, while memory request is not specified
func TestSetZeroRequestMemory(t *testing.T) {
	pod := createPod()
	idx := 0

	rtype := k8sapi.ResourceMemory
	spec := NewContainerResizeSpec(idx)
	amount, err := genMemoryQuantity(128.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtype] = amount

	resizer := &ContainerResizer{}
	resizer.setZeroRequest(pod.Name, &pod.Spec, idx, spec)

	if v, exist := spec.NewRequest[rtype]; !exist {
		t.Error("Failed to set zero request")
	} else {
		fmt.Printf("rtype=%v, v=%++v", rtype, v)
	}
}

// Test Case: resize CPU Capacity, while cpu request is not specified
func TestSetZeroRequestCPU(t *testing.T) {
	pod := createPod()
	idx := 0

	rtype := k8sapi.ResourceCPU
	spec := NewContainerResizeSpec(idx)
	amount, err := genCPUQuantity(1200.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtype] = amount
	fmt.Printf("amount = %++v", amount)

	resizer := &ContainerResizer{}
	resizer.setZeroRequest(pod.Name, &pod.Spec, idx, spec)

	if v, exist := spec.NewRequest[rtype]; !exist {
		t.Error("Failed to set zero request")
	} else {
		fmt.Printf("rtype=%v, v=%++v", rtype, v)
	}
}

// Test Case: resize Memory Capacity, while memory request is already specified;
// In this case, we should not modify the request
func TestSetZeroRequestMemory2(t *testing.T) {
	pod := createPod()
	idx := 0

	rtype := k8sapi.ResourceMemory

	//1. specify memory request
	req := pod.Spec.Containers[idx].Resources.Requests
	amount1, err := genMemoryQuantity(8.0)
	if err != nil {
		t.Errorf("Failed to generate Memory quantity: %v", err)
		return
	}
	req[rtype] = amount1

	//2. set the new Memory capacity
	spec := NewContainerResizeSpec(idx)
	amount, err := genMemoryQuantity(128.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtype] = amount

	resizer := &ContainerResizer{}
	resizer.setZeroRequest(pod.Name, &pod.Spec, idx, spec)

	if _, exist := spec.NewRequest[rtype]; exist {
		t.Errorf("Should not set %v to zero", rtype)
	}
}

// Test Case: resize CPU Capacity, while cpu request is already specified;
// In this case, we should not modify the cpu request
func TestSetZeroRequestCPU2(t *testing.T) {
	pod := createPod()
	idx := 0

	rtype := k8sapi.ResourceCPU

	//1. specify memory request
	req := pod.Spec.Containers[idx].Resources.Requests
	amount1, err := genCPUQuantity(100.0)
	if err != nil {
		t.Errorf("Failed to generate CPU quantity: %v", err)
		return
	}
	req[rtype] = amount1

	//2. set the new Memory capacity
	spec := NewContainerResizeSpec(idx)
	amount, err := genCPUQuantity(1200.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtype] = amount

	//3. update it
	resizer := &ContainerResizer{}
	resizer.setZeroRequest(pod.Name, &pod.Spec, idx, spec)

	//4. check result
	if _, exist := spec.NewRequest[rtype]; exist {
		t.Errorf("Should not set %v to zero", rtype)
	}
}

// Test Case: resize Memory Capacity, while cpu request is already specified;
// In this case, we should only modify the memory request
func TestSetZeroRequestCPUMemory(t *testing.T) {
	pod := createPod()
	idx := 0

	rtypeCPU := k8sapi.ResourceCPU
	rtypeMem := k8sapi.ResourceMemory

	//1. specify memory request
	req := pod.Spec.Containers[idx].Resources.Requests
	amount1, err := genCPUQuantity(100.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	req[rtypeCPU] = amount1

	//2. set the new Memory capacity
	spec := NewContainerResizeSpec(idx)
	amount, err := genMemoryQuantity(1200.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtypeMem] = amount

	//3. update it
	resizer := &ContainerResizer{}
	resizer.setZeroRequest(pod.Name, &pod.Spec, idx, spec)

	//4. check it
	if v, exist := spec.NewRequest[rtypeMem]; !exist {
		t.Error("Failed to set zero request")
	} else {
		fmt.Printf("rtype=%v, v=%++v", rtypeMem, v)
	}

	if len(spec.NewRequest) != 1 {
		t.Errorf("Should only set %v, %d", rtypeMem, len(spec.NewRequest))
	}
}

// Test Case: resize both CPU and Memory Capacity, while neither of them is specified;
// In this case, we should modify both the cpu and memory request to zero
func TestSetZeroRequestCPUMemory2(t *testing.T) {
	pod := createPod()
	idx := 0

	rtypeCPU := k8sapi.ResourceCPU
	rtypeMem := k8sapi.ResourceMemory

	//1. set the new Memory capacity
	spec := NewContainerResizeSpec(idx)
	amount, err := genMemoryQuantity(1200.0)
	if err != nil {
		t.Errorf("Failed to generate memory Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtypeMem] = amount

	//2. set the new CPU capacity
	amount2, err := genCPUQuantity(1000.0)
	if err != nil {
		t.Errorf("Failed to generate cpu Quantity: %v", err)
		return
	}
	spec.NewCapacity[rtypeCPU] = amount2

	//3. update it
	resizer := &ContainerResizer{}
	resizer.setZeroRequest(pod.Name, &pod.Spec, idx, spec)

	//4. check it
	if len(spec.NewRequest) != 2 {
		t.Errorf("Should set %v and %v, %d", rtypeMem, rtypeCPU, len(spec.NewRequest))
	}

	if v, exist := spec.NewRequest[rtypeMem]; !exist {
		t.Error("Failed to set Memory zero request")
	} else {
		fmt.Printf("rtype=%v, v=%++v", rtypeMem, v)
	}

	if v, exist := spec.NewRequest[rtypeCPU]; !exist {
		t.Error("Failed to set CPU zero request")
	} else {
		fmt.Printf("rtype=%v, v=%++v", rtypeCPU, v)
	}
}

func TestBuildResourceListsWithLimits(t *testing.T) {
	actionItem := &proto.ActionItemDTO{
		CurrentComm: mockCommodity(proto.CommodityDTO_VMEM, 10),
		NewComm:     mockCommodity(proto.CommodityDTO_VMEM, 20),
	}
	spec := NewContainerResizeSpec(0)
	resizer := &ContainerResizer{}
	resizer.buildResourceLists(createPod().Name, actionItem, spec)

	assert.Equal(t, 0, len(spec.NewRequest))
	newMemoryCapacity := spec.NewCapacity[k8sapi.ResourceMemory]
	assert.Equal(t, "20Ki", newMemoryCapacity.String())
}

func TestBuildResourceListsWithRequests(t *testing.T) {
	actionItem := &proto.ActionItemDTO{
		CurrentComm: mockCommodity(proto.CommodityDTO_VMEM_REQUEST, 10),
		NewComm:     mockCommodity(proto.CommodityDTO_VMEM_REQUEST, 20),
	}
	spec := NewContainerResizeSpec(0)
	resizer := &ContainerResizer{}
	resizer.buildResourceLists(createPod().Name, actionItem, spec)

	assert.Equal(t, 0, len(spec.NewCapacity))
	newMemoryRequests := spec.NewRequest[k8sapi.ResourceMemory]
	assert.EqualValues(t, "20Ki", newMemoryRequests.String())
}

func TestGenCPUAndMemQuantity(t *testing.T) {
	amount, _ := genCPUQuantity(1.9999)
	assert.Equal(t, "2m", amount.String())
	amount, _ = genMemoryQuantity(1.9999)
	assert.Equal(t, "2Ki", amount.String())
	amount, _ = genCPUQuantity(1.001)
	assert.Equal(t, "1m", amount.String())
	amount, _ = genMemoryQuantity(1.001)
	assert.Equal(t, "1Ki", amount.String())
}

func mockCommodity(commodityType proto.CommodityDTO_CommodityType, capacity float64) *proto.CommodityDTO {
	return &proto.CommodityDTO{
		CommodityType: &commodityType,
		Capacity:      &capacity,
	}
}
