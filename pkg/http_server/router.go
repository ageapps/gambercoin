package http_server

import (
	"net/http"

	"github.com/ageapps/gambercoin/pkg/utils"
	"github.com/gorilla/mux"
)

// Routes arr
type Routes []Route

// MAX_UPLOAD_SIZE contant
// const MAX_UPLOAD_SIZE = 1000 * 1024 // 10 MB

// NewRouter func
func NewRouter(logRequests bool) *mux.Router {

	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes {
		var handler http.Handler

		handler = route.HandlerFunc
		if logRequests {
			handler = Logger(handler, route.Name)
		}
		router.
			Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(handler)
	}

	router.PathPrefix("/").Handler(http.FileServer(http.Dir(utils.GetRootPath() + "/web")))

	return router
}
