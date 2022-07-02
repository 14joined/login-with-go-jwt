package handlers

import (
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
)

type Methods map[string]http.Handler

func (ms Methods) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func(r io.ReadCloser) {
		_, _ = io.Copy(ioutil.Discard, r)
		_ = r.Close()
	}(r.Body)

	if handler, ok := ms[r.Method]; ok {
		if handler == nil {
			http.Error(w, "Internal server error",
				http.StatusInternalServerError)
		} else {
			handler.ServeHTTP(w, r)
		}

		return
	}

	w.Header().Add("Allow", ms.allowedMethods())
	if r.Method != http.MethodOptions {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ms Methods) allowedMethods() string {
	a := make([]string, 0, len(ms))

	for k := range ms {
		a = append(a, k)
	}
	sort.Strings(a)

	return strings.Join(a, ",")
}

