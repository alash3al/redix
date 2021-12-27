package http

import (
	"io"
	"net/http"
	"strconv"

	"github.com/alash3al/redix/internals/manager"
)

// ListenAndServe start the http server
func ListenAndServe(addr string, mngr *manager.Manager) error {
	http.HandleFunc("/dump", func(w http.ResponseWriter, r *http.Request) {
		currentOffset, err := mngr.CurrentOffset()
		if err != nil {
			http.Error(w, "unable to get the current state offset", 500)
			return
		}

		err = mngr.Export(func(size int64) io.Writer {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Disposition", `attachment; filename="dump.rxdb"`)
			w.Header().Set("Content-Length", strconv.Itoa(int(size)))
			w.Header().Set("X-Current-Offset", currentOffset)

			return w
		})

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	return http.ListenAndServe(addr, nil)
}
