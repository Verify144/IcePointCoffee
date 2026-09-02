package server

import (
	"net/http"
	"strconv"
)

// handleCommandHistory 命令历史 API
// GET /api/v1/commands - 列出历史
// DELETE /api/v1/commands - 清空历史
func (s *Server) handleCommandHistory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.cmdHistoryList(w, r)
	case http.MethodDelete:
		s.cmdHistoryClear(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) cmdHistoryList(w http.ResponseWriter, r *http.Request) {
	if s.cmdHistory == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "history not enabled", "")
		return
	}
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}
	list, err := s.cmdHistory.List(limit)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "list failed", err.Error())
		return
	}
	count, _ := s.cmdHistory.Count()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"commands": list,
		"count":    len(list),
		"total":    count,
	})
}

func (s *Server) cmdHistoryClear(w http.ResponseWriter, r *http.Request) {
	if s.cmdHistory == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "history not enabled", "")
		return
	}
	if err := s.cmdHistory.Clear(); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "clear failed", err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "cleared"})
}
