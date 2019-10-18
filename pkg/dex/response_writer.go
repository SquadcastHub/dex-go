package dex

import (
	"net/http"
)

// Response is the dex implementation of a
// http.ResponseWriter interface with metrics enabled
type Response struct {
	w          http.ResponseWriter
	statusCode int
}

// Header method implements the Header method of the
// http.ResponseWriter interface
func (r *Response) Header() http.Header {
	return r.w.Header()
}

// Write method implements the Write method of the
// http.ResponseWriter and io.Writer interface
func (r *Response) Write(b []byte) (int, error) {
	if r.statusCode == 0 {
		r.statusCode = 200
	}
	return r.w.Write(b)
}

// WriteHeader method implements the WriteHeader method of the
// http.ResponseWriter
func (r *Response) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.w.WriteHeader(statusCode)
}
