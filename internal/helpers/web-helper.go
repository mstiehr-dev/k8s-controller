package helpers

import (
	"net/http"
	"golang.org/x/exp/slog"
)

func httpError(w http.ResponseWriter, err error) {
	slog.Error("unable to complete request", "error", err.Error())
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
