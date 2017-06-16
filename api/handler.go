package api

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/smancke/mailigo/mail"
	"io/ioutil"
	"net/http"
	"strings"
)

type Handler struct {
	manager *mail.MailingManager
	router  http.Handler
}

func NewHandler(manager *mail.MailingManager) *Handler {
	h := &Handler{manager: manager}

	r := mux.NewRouter()
	r.HandleFunc("/health", h.Health).Methods("GET")
	r.HandleFunc("/api/mailing/{template}", h.DoMailing).Methods("PUT")

	h.router = r
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	println("foo")
	h.router.ServeHTTP(w, r)
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	fmt.Fprintf(w, "OK\n")
}

func (h *Handler) DoMailing(w http.ResponseWriter, r *http.Request) {
	println("bar")
	vars := mux.Vars(r)
	template, exist := vars["template"]
	if !exist {
		http.Error(w, "missing path variable templates", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "expecting content type json", http.StatusBadRequest)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "body read error", http.StatusBadRequest)
		return
	}

	buff := bytes.NewBuffer(body)

	mailingId, err := h.manager.DoMailing(template, buff)
	if err != nil {
		http.Error(w, "could not create mailing: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", "/api/jobs/"+mailingId)
	w.WriteHeader(201)
	fmt.Fprintf(w, "mailing created: %v\n", mailingId)
}
