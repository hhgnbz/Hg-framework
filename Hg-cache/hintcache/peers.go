package hintcache

// PeerPicker is the interface that must be implemented to locate
// the peer that owns a specific key.
type PeerPicker interface {
	// PeerGetter 用于根据传入的 key 选择相应节点 PeerGetter。
	PeerGetter(key string) (peer PeerGetter, ok bool)
}

// PeerGetter is the interface that must be implemented by a peer.
type PeerGetter interface {
	// Get 方法用于从对应 group 查找缓存值。PeerGetter 就对应于上述流程中的 HTTP 客户端。
	Get(group string, key string) (val []byte, err error)
}
