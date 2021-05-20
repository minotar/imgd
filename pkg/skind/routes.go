package skind

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Requires "uuid" or "username" vars
func SkinPageHandler(storage map[int]map[string]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var uuid string
		vars := mux.Vars(r)

		if username, name_req := vars["username"]; name_req {
			uuid, _ = storage[1][username]
			fmt.Printf("uuid: %+v\n", uuid)
		}

		fmt.Printf("We got the UUID as: %+v", uuid)
		w.WriteHeader(204)
	})
}
