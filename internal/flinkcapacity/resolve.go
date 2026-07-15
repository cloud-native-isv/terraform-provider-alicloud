package flinkcapacity

import "fmt"

func Resolve(input Tree) (Tree, error) {
	tree := cloneTree(input)
	workspaceCapacity := tree.Workspace.AsCapacity()
	if err := workspaceCapacity.Validate(); err != nil {
		return Tree{}, fmt.Errorf("workspace capacity: %w", err)
	}
	if tree.Workspace.Used < 0 {
		return Tree{}, fmt.Errorf("workspace used CU must be non-negative")
	}
	if tree.Workspace.Used > workspaceCapacity.Limit {
		return Tree{}, fmt.Errorf("workspace used CU exceeds its limit")
	}

	switch tree.ChargeType {
	case "PRE":
		if tree.Workspace.TotalFixed() <= 0 {
			return Tree{}, fmt.Errorf("PRE workspace fixed CU total must be greater than zero")
		}
	case "POST":
		if tree.Workspace.TotalFixed() != 0 {
			return Tree{}, fmt.Errorf("POST workspace fixed CU must be zero")
		}
		if tree.Workspace.Limit <= 0 {
			return Tree{}, fmt.Errorf("POST workspace CU limit must be greater than zero")
		}
		// POST has no prepaid fixed-CU component, but its pay-as-you-go limit
		// is still the allocation budget for child guaranteed/request quotas.
		workspaceCapacity.Fixed = workspaceCapacity.Limit
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
		if namespaces[i].Used < 0 {
			return fmt.Errorf("namespace %q used CU must be non-negative", namespaces[i].Name)
		}
		if namespaces[i].Used > namespaces[i].Capacity.Limit {
			return fmt.Errorf("namespace %q used CU exceeds its limit", namespaces[i].Name)
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
		if queues[i].Used < 0 {
			return fmt.Errorf("queue %q/%q used CU must be non-negative", namespaceName, queues[i].Name)
		}
		if queues[i].Used > queues[i].Capacity.Limit {
			return fmt.Errorf("queue %q/%q used CU exceeds its limit", namespaceName, queues[i].Name)
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
		if child.capacity.Limit <= 0 {
			return fmt.Errorf("%s child %q capacity limit must be greater than zero", parentName, child.name)
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
