package http

import (
	"encoding/json"
	"net/http"

	"pr-reviewer/internal/domain"
	"pr-reviewer/internal/service"
)

type teamHandlers struct {
	teams service.TeamService
}

func newTeamHandlers(teams service.TeamService) *teamHandlers {
	return &teamHandlers{teams: teams}
}

type teamMemberDTO struct {
	ID       string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type teamDTO struct {
	Name    string          `json:"team_name"`
	Members []teamMemberDTO `json:"members"`
}

type addTeamRequest struct {
	TeamName string          `json:"team_name"`
	Members  []teamMemberDTO `json:"members"`
}

type addTeamResponse struct {
	Team teamDTO `json:"team"`
}

func (h *teamHandlers) Add(w http.ResponseWriter, r *http.Request) {
	var req addTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeBadRequest(w, "invalid request body")
		return
	}
	if req.TeamName == "" {
		writeBadRequest(w, "team_name is required")
		return
	}

	team := domain.Team{
		Name: req.TeamName,
	}
	for _, m := range req.Members {
		if m.ID == "" || m.Username == "" {
			writeBadRequest(w, "member user_id and username are required")
			return
		}
		team.Members = append(team.Members, domain.User{
			ID:       m.ID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	created, err := h.teams.AddTeam(r.Context(), team)
	if err != nil {
		WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, addTeamResponse{
		Team: toTeamDTO(*created),
	})
}

type getTeamResponse = teamDTO

func (h *teamHandlers) Get(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		writeBadRequest(w, "team_name is required")
		return
	}

	team, err := h.teams.GetTeam(r.Context(), teamName)
	if err != nil {
		WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, toTeamDTO(*team))
}

func toTeamDTO(team domain.Team) teamDTO {
	members := make([]teamMemberDTO, 0, len(team.Members))
	for _, m := range team.Members {
		members = append(members, teamMemberDTO{
			ID:       m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return teamDTO{
		Name:    team.Name,
		Members: members,
	}
}
