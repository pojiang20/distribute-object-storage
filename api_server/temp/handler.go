package temp

import "net/http"

func Handler(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	switch method {
	case http.MethodHead:
		head(w, r)
	case http.MethodPut:
		put(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}
