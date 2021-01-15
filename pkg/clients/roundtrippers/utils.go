package roundtrippers

import "net/http"

func cloneRequest(req *http.Request) *http.Request {
	r := new(http.Request)

	// shallow clone
	*r = *req

	// deep copy headers
	r.Header = make(http.Header)
	for k, v := range req.Header {
		r.Header[k] = v
	}

	return r
}
