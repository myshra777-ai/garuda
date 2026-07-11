package lineage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// EdgeType defines the relationship between two decisions.
type EdgeType int

const (
	EdgeDependsOn EdgeType = iota
	EdgeSupersedes
)

func (et EdgeType) String() string {
	if et == EdgeDependsOn {
		return "depends_on"
	}
	return "supersedes"
}

// Edge represents a directed relationship between two decisions.
type Edge struct {
	From uuid.UUID
	To   uuid.UUID
	Type EdgeType
}

// Graph is a thread-safe, bounded in-memory DAG.
type Graph struct {
	mu          sync.RWMutex
	out         map[uuid.UUID][]Edge
	in          map[uuid.UUID][]Edge
	maxNodes    int
	nodeCount   int
}

func NewGraph(maxNodes int) *Graph {
	return &Graph{
		out:       make(map[uuid.UUID][]Edge),
		in:        make(map[uuid.UUID][]Edge),
		maxNodes:  maxNodes,
		nodeCount: 0,
	}
}

func (g *Graph) AddEdge(from, to uuid.UUID, etype EdgeType) error {
	if from == to {
		return errors.New("self‑edge is not allowed")
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	if g.edgeExists(from, to, etype) {
		return nil
	}

	// Calculate unique nodes after adding this edge
	nodeSet := make(map[uuid.UUID]bool)
	for fromID := range g.out {
		nodeSet[fromID] = true
	}
	for toID := range g.in {
		nodeSet[toID] = true
	}
	nodeSet[from] = true
	nodeSet[to] = true

	if g.maxNodes > 0 && len(nodeSet) > g.maxNodes {
		return fmt.Errorf("graph would exceed maximum node limit (%d)", g.maxNodes)
	}

	if g.wouldCreateCycle(from, to) {
		return fmt.Errorf("adding edge %v → %v would create a cycle", from, to)
	}

	g.out[from] = append(g.out[from], Edge{From: from, To: to, Type: etype})
	g.in[to] = append(g.in[to], Edge{From: from, To: to, Type: etype})

	g.nodeCount = len(nodeSet)

	return nil
}

func (g *Graph) edgeExists(from, to uuid.UUID, etype EdgeType) bool {
	for _, e := range g.out[from] {
		if e.To == to && e.Type == etype {
			return true
		}
	}
	return false
}

func (g *Graph) wouldCreateCycle(from, to uuid.UUID) bool {
	visited := make(map[uuid.UUID]bool)
	return g.dfs(to, from, visited)
}

func (g *Graph) dfs(current, target uuid.UUID, visited map[uuid.UUID]bool) bool {
	if current == target {
		return true
	}
	visited[current] = true
	for _, edge := range g.out[current] {
		if !visited[edge.To] {
			if g.dfs(edge.To, target, visited) {
				return true
			}
		}
	}
	return false
}

func (g *Graph) GetChildren(id uuid.UUID) []uuid.UUID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var children []uuid.UUID
	for _, edge := range g.in[id] {
		if edge.Type == EdgeDependsOn {
			children = append(children, edge.From)
		}
	}
	return children
}

func (g *Graph) GetParents(id uuid.UUID) []uuid.UUID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	var parents []uuid.UUID
	for _, edge := range g.out[id] {
		if edge.Type == EdgeDependsOn {
			parents = append(parents, edge.To)
		}
	}
	return parents
}

func (g *Graph) GetImpactSet(id uuid.UUID) []uuid.UUID {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[uuid.UUID]bool)
	var impact []uuid.UUID
	g.collectImpact(id, visited, &impact)
	return impact
}

func (g *Graph) collectImpact(current uuid.UUID, visited map[uuid.UUID]bool, impact *[]uuid.UUID) {
	for _, child := range g.GetChildren(current) {
		if !visited[child] {
			visited[child] = true
			*impact = append(*impact, child)
			g.collectImpact(child, visited, impact)
		}
	}
}

func (g *Graph) RemoveDecision(id uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()

	for _, edge := range g.out[id] {
		inEdges := g.in[edge.To]
		for i, e := range inEdges {
			if e.From == id && e.Type == edge.Type {
				g.in[edge.To] = append(inEdges[:i], inEdges[i+1:]...)
				break
			}
		}
	}
	delete(g.out, id)

	for _, edge := range g.in[id] {
		outEdges := g.out[edge.From]
		for i, e := range outEdges {
			if e.To == id && e.Type == edge.Type {
				g.out[edge.From] = append(outEdges[:i], outEdges[i+1:]...)
				break
			}
		}
	}
	delete(g.in, id)

	g.nodeCount = len(g.out) + len(g.in) - g.overlapCount()
	return nil
}

func (g *Graph) overlapCount() int {
	count := 0
	for id := range g.out {
		if _, exists := g.in[id]; exists {
			count++
		}
	}
	return count
}

func (g *Graph) ValidateIntegrity() error {
	g.mu.RLock()
	defer g.mu.RUnlock()

	visited := make(map[uuid.UUID]bool)
	recStack := make(map[uuid.UUID]bool)

	var dfsCycle func(id uuid.UUID) bool
	dfsCycle = func(id uuid.UUID) bool {
		visited[id] = true
		recStack[id] = true
		for _, edge := range g.out[id] {
			if !visited[edge.To] {
				if dfsCycle(edge.To) {
					return true
				}
			} else if recStack[edge.To] {
				return true
			}
		}
		recStack[id] = false
		return false
	}

	for id := range g.out {
		if !visited[id] {
			if dfsCycle(id) {
				return fmt.Errorf("cycle detected in graph")
			}
		}
	}
	return nil
}
