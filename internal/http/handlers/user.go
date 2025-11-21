package handlers

import "net/http"

type UserHandlers struct{}

func (h *UserHandlers) SetIsActive(w http.ResponseWriter, r *http.Request) {}
func (h *UserHandlers) GetReview(w http.ResponseWriter, r *http.Request)   {}
