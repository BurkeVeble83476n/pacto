package graph

// ChangeType indicates the kind of dependency graph change.
type ChangeType string

const (
	AddedNode      ChangeType = "added"
	RemovedNode    ChangeType = "removed"
	VersionChanged ChangeType = "version_changed"
)

// GraphChange represents a single change in the dependency graph.
type GraphChange struct {
	Name       string     `json:"name"`
	ChangeType ChangeType `json:"changeType"`
	OldVersion string     `json:"oldVersion,omitempty"`
	NewVersion string     `json:"newVersion,omitempty"`
}

// DiffNode represents a node in the diff tree, carrying its own change
// (if any) and the changes in its subtree.
type DiffNode struct {
	Name     string       `json:"name"`
	Version  string       `json:"version,omitempty"`
	Change   *GraphChange `json:"change,omitempty"`
	Children []DiffNode   `json:"children,omitempty"`
}

// GraphDiff holds the result of comparing two dependency graphs.
type GraphDiff struct {
	Root    DiffNode      `json:"root"`
	Changes []GraphChange `json:"changes,omitempty"`
}

// DiffGraphs compares two resolved dependency graphs and returns
// a GraphDiff describing added, removed, and version-changed nodes.
// Either old or new may be nil (representing an empty graph).
func DiffGraphs(old, new *Result) *GraphDiff {
	oldVersions := flattenVersions(old)
	newVersions := flattenVersions(new)

	var changes []GraphChange
	for name, newVer := range newVersions {
		oldVer, exists := oldVersions[name]
		if !exists {
			changes = append(changes, GraphChange{Name: name, ChangeType: AddedNode, NewVersion: newVer})
		} else if oldVer != newVer {
			changes = append(changes, GraphChange{Name: name, ChangeType: VersionChanged, OldVersion: oldVer, NewVersion: newVer})
		}
	}
	for name, oldVer := range oldVersions {
		if _, exists := newVersions[name]; !exists {
			changes = append(changes, GraphChange{Name: name, ChangeType: RemovedNode, OldVersion: oldVer})
		}
	}

	sortChanges(changes)

	// Build diff tree from the new graph structure, annotating changed nodes.
	changeMap := map[string]*GraphChange{}
	for i := range changes {
		changeMap[changes[i].Name] = &changes[i]
	}

	var root DiffNode
	if new != nil && new.Root != nil {
		root = buildDiffTree(new.Root, oldVersions, changeMap, map[string]bool{})
		// Append removed nodes as children of root (they don't exist in new graph).
		for i := range changes {
			if changes[i].ChangeType == RemovedNode {
				root.Children = append(root.Children, DiffNode{
					Name:    changes[i].Name,
					Version: changes[i].OldVersion,
					Change:  &changes[i],
				})
			}
		}
	} else if old != nil && old.Root != nil {
		root = buildRemovedTree(old.Root, changeMap, map[string]bool{})
	}

	return &GraphDiff{Root: root, Changes: changes}
}

// flattenVersions collects all unique dependency name→version mappings
// from the graph (excluding the root).
func flattenVersions(r *Result) map[string]string {
	if r == nil || r.Root == nil {
		return map[string]string{}
	}
	versions := map[string]string{}
	flattenNode(r.Root, versions)
	return versions
}

func flattenNode(node *Node, versions map[string]string) {
	if node == nil {
		return
	}
	for _, edge := range node.Dependencies {
		if edge.Node == nil {
			continue
		}
		if _, seen := versions[edge.Node.Name]; seen {
			continue
		}
		versions[edge.Node.Name] = edge.Node.Version
		if !edge.Shared {
			flattenNode(edge.Node, versions)
		}
	}
}

// buildDiffTree recursively builds a DiffNode tree from the new graph,
// annotating nodes that have changes.
func buildDiffTree(node *Node, oldVersions map[string]string, changeMap map[string]*GraphChange, visited map[string]bool) DiffNode {
	dn := DiffNode{Name: node.Name, Version: node.Version}

	for _, edge := range node.Dependencies {
		if edge.Node == nil {
			continue
		}
		child := DiffNode{
			Name:    edge.Node.Name,
			Version: edge.Node.Version,
			Change:  changeMap[edge.Node.Name],
		}
		if !edge.Shared && !visited[edge.Node.Name] {
			visited[edge.Node.Name] = true
			child.Children = buildDiffTree(edge.Node, oldVersions, changeMap, visited).Children
		}
		dn.Children = append(dn.Children, child)
	}

	// Append removed nodes that were direct children in old graph but missing in new.
	// These are captured in the flat changes list; we add them at this level.
	return dn
}

// buildRemovedTree builds a diff tree showing all nodes as removed
// (used when the new graph is nil/empty).
func buildRemovedTree(node *Node, changeMap map[string]*GraphChange, visited map[string]bool) DiffNode {
	dn := DiffNode{Name: node.Name, Version: node.Version}

	for _, edge := range node.Dependencies {
		if edge.Node == nil {
			continue
		}
		child := DiffNode{
			Name:    edge.Node.Name,
			Version: edge.Node.Version,
			Change:  changeMap[edge.Node.Name],
		}
		if !edge.Shared && !visited[edge.Node.Name] {
			visited[edge.Node.Name] = true
			child.Children = buildRemovedTree(edge.Node, changeMap, visited).Children
		}
		dn.Children = append(dn.Children, child)
	}
	return dn
}

func sortChanges(changes []GraphChange) {
	for i := 0; i < len(changes); i++ {
		for j := i + 1; j < len(changes); j++ {
			if changes[i].Name > changes[j].Name {
				changes[i], changes[j] = changes[j], changes[i]
			}
		}
	}
}
