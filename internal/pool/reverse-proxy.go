package pool

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/tehrelt/eginx/internal/router"
)

const (
	forwardedBy = "eginx"
)

type reverseProxy struct {
	*httputil.ReverseProxy
}

func newRP(target *url.URL) *reverseProxy {
	return &reverseProxy{
		ReverseProxy: &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = target.Scheme
				req.URL.Host = target.Host
				req.URL.Path = target.Path + req.URL.Path
			},
			ModifyResponse: func(resp *http.Response) error {
				resp.Header.Set("X-Forwarded-For", target.String())
				return nil
			},
			ErrorHandler: func(w http.ResponseWriter, req *http.Request, err error) {
				slog.Debug("proxy error", slog.Any("error", err))

				resp := router.NewError(errors.New("bad gateway"), http.StatusBadGateway)

				w.WriteHeader(resp.Code)
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Forwarded-For", target.String())

				if err := json.NewEncoder(w).Encode(resp); err != nil {
					slog.Error("failed to encode response json", slog.Any("error", err), slog.Any("resp", resp))
					return
				}
			},
		},
	}
}
