package registry

import "encoding/json"

// NodeInfo contains information on a single node
type NodeInfo struct {
	FQDN string
	Port int
}

// NodeInfoFromBytes unmarshalls the []byte into a NodeInfo and returns a reference
func NodeInfoFromBytes(data []byte) (*NodeInfo, error) {
	n := NodeInfo{}
	err := json.Unmarshal(data, &n)

	return &n, err
}

// A Register contains information of all registered nodes of a cluster
type Register struct {
	Nodes []*NodeInfo `json:"nodes"`
}

// RegisterFromBytes unmarshalls the bytes into a Register
func RegisterFromBytes(data []byte) (*Register, error) {
	r := Register{}

	err := json.Unmarshal(data, &r)
	return &r, err
}

// AddNode adds a node to a register
func (r *Register) AddNode(node *NodeInfo) {
	if r.nodeExists(node) {
		return
	}

	r.Nodes = append(r.Nodes, node)
}

func (r *Register) nodeExists(node *NodeInfo) bool {
	for _, n := range r.Nodes {
		if n.Port == node.Port && n.FQDN == node.FQDN {
			return true
		}
	}

	return false
}

// Bytes marshalls the register
func (r *Register) Bytes() ([]byte, error) {
	return json.Marshal(r)
}
