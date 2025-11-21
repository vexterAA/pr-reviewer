package handlers

import "net/http"

type PullRequestHandlers struct{}

func (h *PullRequestHandlers) Create(w http.ResponseWriter, r *http.Request)   {}
func (h *PullRequestHandlers) Merge(w http.ResponseWriter, r *http.Request)    {}
func (h *PullRequestHandlers) Reassign(w http.ResponseWriter, r *http.Request) {}
