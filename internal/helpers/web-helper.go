package helpers

import (
	"golang.org/x/exp/slog"
	"net/http"
)

func httpError(w http.ResponseWriter, err error) {
	slog.Error("unable to complete request", "error", err.Error())
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(err.Error()))
}
