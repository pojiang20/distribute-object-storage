package objects

import (
	"net/http"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	m := r.Method

	if m == http.MethodPut {
		put(w, r)
	} else if m == http.MethodGet {
		get(w, r)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
