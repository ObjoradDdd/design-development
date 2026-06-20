package middleware

import "net/http"

func keyGen(r *http.Request, comr string) string {
	return r.Method + "|" + r.URL.RequestURI() + "|" + comr
}