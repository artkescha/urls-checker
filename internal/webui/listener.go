package webui

import (
	"net/http"
)

func StartListener(addr, username, password string) error {
	h := &Handler{
		username: username,
		password: password,
	}
	h.InitSessions()
	err := h.InitTemplates()
	if err != nil {
		return err
	}
	http.Handle("/static/", http.FileServer(http.FS(staticContent)))
	http.Handle("/", h)
	return http.ListenAndServe(addr, nil)
}
