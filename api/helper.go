package api

import (
  "net/http"
)

func notLoggedIn(w http.ResponseWriter) {
  w.WriteHeader(http.StatusUnauthorized)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusNotFound)
}

func InternalError(w http.ResponseWriter, r *http.Request) {
  w.WriteHeader(http.StatusInternalServerError)
}
