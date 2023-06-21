package hint

import "strings"

// HTTP请求的路径恰好是由/分隔的多段构成的，因此，每一段可以作为前缀树的一个节点。
// 通过树结构查询，如果中间某一层的节点都不满足条件，那么就说明没有匹配到的路由，查询结束。
//
// 参数匹配":"，例如 /p/:lang/doc，可以匹配 /p/c/doc 和 /p/go/doc。
// 通配"*"，例如 /static/*filepath，可以匹配/static/fav.ico，也可以匹配/static/js/jQuery.js，这种模式常用于静态服务器，能够递归地匹配子路径。

type trieNode struct {
	pattern    string      // current full pattern of router e.g. /p/:lang (not nil when the path fulled,"bool end" param)
	curPattern string      // current part of full pattern e.g. /:lang
	children   []*trieNode // child node e.g. [doc,info]
	isWild     bool        // true when pattern contains ":" or "*"
}

// get the first child that the pattern matched tn.curPattern
func (tn *trieNode) matchChild(curPattern string) *trieNode {
	for _, c := range tn.children {
		if c.curPattern == curPattern || c.isWild {
			return c
		}
	}
	return nil
}

// get all child that the pattern matched tn.curPattern
func (tn *trieNode) matchChildren(curPattern string) []*trieNode {
	children := make([]*trieNode, 0)
	for _, c := range tn.children {
		if c.curPattern == curPattern || c.isWild {
			children = append(children, c)
		}
	}
	return children
}

// insert pattern
func (tn *trieNode) insert(pattern string, parts []string, depth int) {
	// tn.pattern not nil when the path fulled
	if depth == len(parts) {
		tn.pattern = pattern
		return
	}
	curPattern := parts[depth]
	child := tn.matchChild(curPattern)
	// insert when child not exist
	if child == nil {
		child = &trieNode{curPattern: curPattern, isWild: curPattern[0] == ':' || curPattern[0] == '*'}
		tn.children = append(tn.children, child)
	}
	child.insert(pattern, parts, depth+1)
}

// search pattern
func (tn *trieNode) search(parts []string, depth int) *trieNode {
	// exit when matched "*" prefix or matched fail(pattern not exist) or depth reached the end of parts(match succeed)
	if len(parts) == depth || strings.HasPrefix(tn.curPattern, "*") {
		if tn.pattern == "" {
			return nil
		}
		return tn
	}
	curPattern := parts[depth]
	children := tn.matchChildren(curPattern)
	for _, child := range children {
		res := child.search(parts, depth+1)
		if res != nil {
			return res
		}
	}
	return nil
}
