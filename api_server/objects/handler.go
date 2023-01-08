package objects

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method
	switch m {
	case http.MethodPut:
		put(w, r)
	case http.MethodPost:
		post(w, r)
	case http.MethodGet:
		get(w, r)
	case http.MethodDelete:
		del(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
	return
}
