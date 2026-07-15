package flinkcapacity

import (
	"fmt"
	"math"
)

func Resolve(input Tree) (Tree, error) {
	tree := cloneTree(input)
	workspaceCapacity := tree.Workspace.AsCapacity()
	if err := workspaceCapacity.Validate(); err != nil {
		return Tree{}, fmt.Errorf("workspace capacity: %w", err)
	}
	if err := validateUsedCU("workspace", tree.Workspace.Used, workspaceCapacity.Limit); err != nil {
		return Tree{}, err
	}
	if tree.Workspace.HA {
		if tree.Workspace.CrossZoneFixedCU <= 0 {
			return Tree{}, fmt.Errorf("HA workspace cross-zone fixed CU must be greater than zero")
		}
	} else if tree.Workspace.CrossZoneFixedCU != 0 {
		return Tree{}, fmt.Errorf("non-HA workspace cannot configure cross-zone fixed CU")
	}

	switch tree.ChargeType {
	case "PRE":
		if tree.Workspace.TotalFixed() <= 0 {
			return Tree{}, fmt.Errorf("PRE workspace fixed CU total must be greater than zero")
		}
	case "POST":
		if tree.Workspace.HA {
			return Tree{}, fmt.Errorf("POST workspace cannot use high availability")
		}
		if tree.Workspace.TotalFixed() != 0 {
			return Tree{}, fmt.Errorf("POST workspace fixed CU must be zero")
		}
		if tree.Workspace.Limit <= 0 {
			return Tree{}, fmt.Errorf("POST workspace CU limit must be greater than zero")
		}
	default:
		return Tree{}, fmt.Errorf("unsupported charge type %q", tree.ChargeType)
	}

	if err := resolveNamespaces(workspaceCapacity, tree.Namespaces); err != nil {
		return Tree{}, err
	}
	return tree, nil
}

func resolveNamespaces(parent Capacity, namespaces []Namespace) error {
	if len(namespaces) == 0 {
		return fmt.Errorf("workspace must declare at least one namespace")
	}

	children := make([]namedCapacity, len(namespaces))
	for i := range namespaces {
		children[i] = namedCapacity{name: namespaces[i].Name, capacity: namespaces[i].Capacity}
	}
	if err := resolveChildren("workspace", parent, children); err != nil {
		return err
	}
	for i := range namespaces {
		namespaces[i].Capacity = children[i].capacity
		if err := validateUsedCU(fmt.Sprintf("namespace %q", namespaces[i].Name), namespaces[i].Used, namespaces[i].Capacity.Limit); err != nil {
			return err
		}
		if err := resolveQueues(namespaces[i].Name, *namespaces[i].Capacity, namespaces[i].Queues); err != nil {
			return err
		}
	}
	return nil
}

func resolveQueues(namespaceName string, parent Capacity, queues []Queue) error {
	if len(queues) == 0 {
		return fmt.Errorf("namespace %q must declare at least one queue", namespaceName)
	}

	children := make([]namedCapacity, len(queues))
	for i := range queues {
		children[i] = namedCapacity{name: queues[i].Name, capacity: queues[i].Capacity}
	}
	if err := resolveChildren("namespace "+namespaceName, parent, children); err != nil {
		return err
	}
	for i := range queues {
		queues[i].Capacity = children[i].capacity
		if err := validateUsedCU(fmt.Sprintf("queue %q/%q", namespaceName, queues[i].Name), queues[i].Used, queues[i].Capacity.Limit); err != nil {
			return err
		}
	}
	return nil
}

type namedCapacity struct {
	name     string
	capacity *Capacity
}

func resolveChildren(parentName string, parent Capacity, children []namedCapacity) error {
	seen := make(map[string]struct{}, len(children))
	remainderIndex := -1
	var explicitFixed, explicitLimit CU

	for i := range children {
		child := &children[i]
		if child.name == "" {
			return fmt.Errorf("%s child name must not be empty", parentName)
		}
		if _, exists := seen[child.name]; exists {
			return fmt.Errorf("%s has duplicate child name %q", parentName, child.name)
		}
		seen[child.name] = struct{}{}

		if child.capacity == nil {
			if remainderIndex >= 0 {
				return fmt.Errorf("%s may have at most one child without an explicit capacity", parentName)
			}
			remainderIndex = i
			continue
		}
		if err := child.capacity.Validate(); err != nil {
			return fmt.Errorf("%s child %q: %w", parentName, child.name, err)
		}
		explicitFixed += child.capacity.Fixed
		explicitLimit += child.capacity.Limit
	}

	if explicitFixed > parent.Fixed {
		return fmt.Errorf("%s child fixed CU %v exceeds parent fixed CU %v", parentName, explicitFixed.Float64(), parent.Fixed.Float64())
	}
	if explicitLimit > parent.Limit {
		return fmt.Errorf("%s child CU limit %v exceeds parent limit %v", parentName, explicitLimit.Float64(), parent.Limit.Float64())
	}

	if remainderIndex >= 0 {
		remainder := Capacity{
			Fixed: parent.Fixed - explicitFixed,
			Limit: parent.Limit - explicitLimit,
		}
		if remainder.Limit <= 0 {
			return fmt.Errorf("%s remainder child capacity limit must be greater than zero", parentName)
		}
		children[remainderIndex].capacity = &remainder
	}
	return nil
}

func validateUsedCU(name string, used float64, limit CU) error {
	if math.IsNaN(used) || math.IsInf(used, 0) || used < 0 {
		return fmt.Errorf("%s used CU must be a finite non-negative value", name)
	}
	if used > limit.Float64() {
		return fmt.Errorf("%s used CU exceeds its limit", name)
	}
	return nil
}

func cloneTree(input Tree) Tree {
	result := input
	result.Namespaces = make([]Namespace, len(input.Namespaces))
	for i := range input.Namespaces {
		result.Namespaces[i] = input.Namespaces[i]
		result.Namespaces[i].Capacity = cloneCapacity(input.Namespaces[i].Capacity)
		result.Namespaces[i].Queues = make([]Queue, len(input.Namespaces[i].Queues))
		for j := range input.Namespaces[i].Queues {
			result.Namespaces[i].Queues[j] = input.Namespaces[i].Queues[j]
			result.Namespaces[i].Queues[j].Capacity = cloneCapacity(input.Namespaces[i].Queues[j].Capacity)
		}
	}
	return result
}

func cloneCapacity(input *Capacity) *Capacity {
	if input == nil {
		return nil
	}
	result := *input
	return &result
}
