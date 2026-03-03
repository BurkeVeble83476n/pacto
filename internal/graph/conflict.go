package graph

import "fmt"

// Conflict represents a version conflict where the same service
// is required at incompatible versions.
type Conflict struct {
	Name     string   `json:"name"`
	Versions []string `json:"versions"`
}

// detectConflicts walks the graph and finds services that appear
// with multiple different versions.
func detectConflicts(root *Node) []Conflict {
	versions := map[string]map[string]bool{}
	collectVersions(root, versions)

	var conflicts []Conflict
	for name, vs := range versions {
		if len(vs) > 1 {
			var list []string
			for v := range vs {
				list = append(list, fmt.Sprintf("%s@%s", name, v))
			}
			conflicts = append(conflicts, Conflict{Name: name, Versions: list})
		}
	}
	return conflicts
}

// collectVersions recursively collects service name → version mappings.
func collectVersions(node *Node, versions map[string]map[string]bool) {
	if node == nil {
		return
	}
	for _, edge := range node.Dependencies {
		if edge.Node != nil {
			if versions[edge.Node.Name] == nil {
				versions[edge.Node.Name] = map[string]bool{}
			}
			versions[edge.Node.Name][edge.Node.Version] = true
			collectVersions(edge.Node, versions)
		}
	}
}
