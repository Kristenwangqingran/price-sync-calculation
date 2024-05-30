package router

import (
	"strings"

	"git.garena.com/shopee/core-server/price/price-sync-calculation/internal/util/miscellaneousinner/router/pkg"

	"git.garena.com/shopee/platform/golang_splib/desc/http"
)

type router struct {
	reversedTree  pkg.ReversedTree
	patternToRule map[http.Rule]http.Rule
}

func New(rules []http.Rule) http.Router {
	r := router{
		reversedTree:  pkg.NewGinReverseRouter(),
		patternToRule: make(map[http.Rule]http.Rule),
	}

	for _, v := range rules {
		v.Method = strings.ToUpper(v.Method)
		r.reversedTree.Insert(v)
		r.patternToRule[http.Rule{Method: v.Method, Path: v.Path}] = v
	}

	return &r
}

func (r *router) lookup(method, path string) (http.Rule, bool) {
	rule, exist := r.reversedTree.Lookup(strings.ToUpper(method), path)
	if exist {
		return r.patternToRule[rule], true
	}

	const MethodAll = "ALL"
	rule, exist = r.reversedTree.Lookup(MethodAll, path)

	return r.patternToRule[rule], exist
}

func (r *router) Lookup(method, path string) (http.Rule, bool) {
	rule, ok := r.lookup(method, path)
	if !ok {
		n := len(path)
		trailing := n > 1 && path[n-1] == '/'
		if trailing {
			rule, ok = r.lookup(method, path[:n-1])
		}
	}

	return rule, ok
}
