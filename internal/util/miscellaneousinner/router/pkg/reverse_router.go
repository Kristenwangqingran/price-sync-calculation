package pkg

import (
	"net/url"

	"git.garena.com/shopee/platform/golang_splib/desc/http"
)

type ReversedRoute = http.Rule

type ReversedTree interface {
	Insert(r ReversedRoute)
	Lookup(method string, path string) (ReversedRoute, bool)
}

var _ ReversedTree = new(ginReverseRoute)

var fakedHandler HandlerFunc = func(r *Context) {
}

// nolint:godox //check whether need add options
type ginReverseRoute struct {
	trees       methodTrees
	maxParams   uint16
	maxSections uint16

	// If enabled, the url.RawPath will be used to find parameters.
	UseRawPath bool // default false
	// If true, the path value will be unescaped.
	// If UseRawPath is false (by default), the UnescapePathValues effectively is true,
	// as url.Path gonna be used, which is already unescaped.
	UnescapePathValues bool // default false

	// RemoveExtraSlash a parameter can be parsed from the Path even with extra slashes.
	// See the PR #1817 and issue #1644
	RemoveExtraSlash bool // default true
}

func NewGinReverseRouter() ReversedTree {
	return &ginReverseRoute{
		trees:              make(methodTrees, 0, 9),
		UseRawPath:         false,
		RemoveExtraSlash:   false,
		UnescapePathValues: true,
	}
}

func (g *ginReverseRoute) Insert(r ReversedRoute) {
	// reference https://github.com/gin-gonic/gin/blob/master/gin.go#L292
	root := g.trees.get(r.Method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"

		g.trees = append(g.trees, methodTree{
			method: r.Method,
			root:   root,
		},
		)
	}

	root.addRoute(r.Path, HandlersChain{fakedHandler})
	// Update maxParams
	if paramsCount := countParams(r.Path); paramsCount > g.maxParams {
		g.maxParams = paramsCount
	}

	if sectionsCount := countSections(r.Path); sectionsCount > g.maxSections {
		g.maxSections = sectionsCount
	}
}

func (g *ginReverseRoute) Lookup(method, path string) (ReversedRoute, bool) {
	httpMethod := method
	rPath := path
	unescape := false

	if g.UseRawPath {
		URL, err := url.Parse(path)
		if err == nil && len(URL.RawPath) > 0 {
			rPath = URL.RawPath
			unescape = g.UnescapePathValues
		}
	}

	if g.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	skippedNodes := make([]skippedNode, 0, g.maxSections)
	// Find root of the tree for the given HTTP method
	t := g.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}

		root := t[i].root
		// Find route in tree
		value := root.getValue(rPath, nil, &skippedNodes, unescape)
		if value.handlers != nil {
			return ReversedRoute{
				Path:   value.fullPath,
				Method: method,
			}, true
		}

		break
	}

	return ReversedRoute{}, false
}
