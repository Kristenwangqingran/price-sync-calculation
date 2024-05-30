package pkg

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(r *Context)

// HandlersChain defines a HandlerFunc array.
type HandlersChain []HandlerFunc

type Context struct{}
