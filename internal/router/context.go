package router

import "net/http"

type HandlerFn func(w http.ResponseWriter, r *http.Request) error
