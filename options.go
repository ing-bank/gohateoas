package gohateoas

import "net/http"

// LinkInfo represents a link to a resource.
type LinkInfo struct {
	Method  string `json:"method"`
	Href    string `json:"href"`
	Comment string `json:"comment"`
}

// LinkOption is used to register links in a LinkRegistry. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
type LinkOption func(map[string]LinkInfo)

// Custom allows you to define a custom action and info. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Custom(action string, info LinkInfo) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry[action] = info
	}
}

// Self Adds the self url of an object to the type, probably an url with an id. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Self(href string, comment string) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry["self"] = LinkInfo{
			Method:  http.MethodGet,
			Href:    href,
			Comment: comment,
		}
	}
}

// Index Adds a general Index route to the type. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Index(href string, comment string) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry["index"] = LinkInfo{
			Method:  http.MethodGet,
			Href:    href,
			Comment: comment,
		}
	}
}

// Post Adds a general POST route to the LinkRegistry. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Post(href string, comment string) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry["post"] = LinkInfo{
			Method:  http.MethodPost,
			Href:    href,
			Comment: comment,
		}
	}
}

// Put Adds a general Put route to the LinkRegistry. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Put(href string, comment string) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry["put"] = LinkInfo{
			Method:  http.MethodPut,
			Href:    href,
			Comment: comment,
		}
	}
}

// Patch Adds a general Patch route to the LinkRegistry. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Patch(url string, comment string) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry["patch"] = LinkInfo{
			Method:  http.MethodPatch,
			Href:    url,
			Comment: comment,
		}
	}
}

// Delete Adds a general Delete route to the LinkRegistry. Urls may contain
// replaceable tokens like {id} or {name}. These tokens will be replaced by
// the values of the corresponding json fields in the struct.
func Delete(url string, comment string) LinkOption {
	return func(registry map[string]LinkInfo) {
		registry["delete"] = LinkInfo{
			Method:  http.MethodDelete,
			Href:    url,
			Comment: comment,
		}
	}
}
