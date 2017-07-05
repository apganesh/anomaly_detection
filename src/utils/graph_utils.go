package anomaly

/////////////////////////////////////////////////////////////////////////////
// Simple graph represetation  using adjacency list
/////////////////////////////////////////////////////////////////////////////
type Graph map[uint32]map[uint32]bool

/////////////////////////////////////////////////////////////////////////////
// Graph Utilities
/////////////////////////////////////////////////////////////////////////////

// Add Vertex
func (g *Graph) addVertex(id uint32) {
	edges := (*g)[id]
	if edges == nil {
		edges = make(map[uint32]bool)
		(*g)[id] = edges
	}
}

// Add an edge (single direction)
func (g *Graph) addEdge(from, to uint32) {
	edges := (*g)[from]
	if edges == nil {
		edges = make(map[uint32]bool)
		(*g)[from] = edges
	}
	edges[to] = true
}

// Add both directional edges between two vertices
func (g *Graph) AddUndirectedEdge(id1, id2 uint32) {
	g.addEdge(id1, id2)
	g.addEdge(id2, id1)
}

// remove and edge, but not the vertices
func (g *Graph) remEdge(from, to uint32) {
	edges := (*g)[from]
	if edges == nil {
		return
	}
	delete((*g)[from], to)
}

// Remove both directional edges between two vertices
func (g *Graph) RemUndirectedEdge(id1, id2 uint32) {
	if g.hasEdge(id1, id2) {
		g.remEdge(id1, id2)
		g.remEdge(id2, id1)
	}
}

func (g *Graph) hasEdge(from, to uint32) bool {
	return (*g)[from][to]
}

func (g *Graph) getFriends_BFS(id uint32, degree uint32) []uint32 {

	type Node struct {
		id     uint32
		degree uint32
	}
	var res []uint32

	var q []Node
	q = append(q, Node{id, degree})

	vis := make(map[uint32]bool)
	vis[id] = true

	for len(q) > 0 {
		n := q[0]
		q = q[1:]
		if n.degree == 0 {
			if _, ok := vis[n.id]; !ok {
				vis[n.id] = true
				res = append(res, n.id)
			}
			continue
		}

		for fid, _ := range (*g)[n.id] {
			if _, ok := vis[fid]; ok {
				continue
			}
			q = append(q, Node{fid, n.degree - 1})
			vis[fid] = true
			res = append(res, fid)
		}
	}
	return res
}

// Public API to get friends
func (g *Graph) GetFriends(id uint32, degree uint32) []uint32 {
	res := g.getFriends_BFS(id, degree)
	return res
}
