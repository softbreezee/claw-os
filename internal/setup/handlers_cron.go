package setup

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/softbreezee/claw-os/internal/store"
)

// --- Cron Jobs ---
//
// All four handlers below back onto s.cronStore (the unified store
// shared with the agent tools and scheduler). Pre-v0.2.x the same
// endpoints persisted to cfg.CronJobs in pawnix.json, which created
// the "agent-created jobs invisible in UI" bug — this file is the
// fix's API surface.

// jobToJSON renders a CronJobRecord into the wire format the web
// dashboard expects. Kept in one place so all four handlers stay in
// lockstep.
func jobToJSON(j store.CronJobRecord) map[string]any {
	out := map[string]any{
		"id":        j.ID,
		"name":      j.Name,
		"type":      j.Type,
		"schedule":  j.Schedule,
		"agentId":   j.AgentID,
		"channel":   j.Channel,
		"chatId":    j.ChatID,
		"message":   j.Message,
		"enabled":   j.Enabled,
		"createdAt": j.CreatedAt.Format(time.RFC3339),
	}
	if j.LastRun != nil {
		out["lastRun"] = j.LastRun.Format(time.RFC3339)
	}
	if j.NextRun != nil {
		out["nextRun"] = j.NextRun.Format(time.RFC3339)
	}
	return out
}

func (s *Server) handleListCronJobs(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusOK, []any{})
		return
	}
	jobs, err := s.cronStore.ListCronJobs(r.Context(), s.cronTenantID)
	if err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	out := make([]map[string]any, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, jobToJSON(j))
	}
	jsonResponse(w, http.StatusOK, out)
}

func (s *Server) handleCreateCronJob(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "cron store not initialised"})
		return
	}
	var req struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Schedule string `json:"schedule"`
		AgentID  string `json:"agentId"`
		Channel  string `json:"channel"`
		ChatID   string `json:"chatId"`
		Message  string `json:"message"`
		Enabled  *bool  `json:"enabled,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	if req.Name == "" || req.Schedule == "" {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "name and schedule are required"})
		return
	}
	if req.Type == "" {
		req.Type = "cron"
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	now := time.Now()
	rec := &store.CronJobRecord{
		ID:        newCronJobID(),
		TenantID:  s.cronTenantID,
		AgentID:   req.AgentID,
		Name:      req.Name,
		Type:      req.Type,
		Schedule:  req.Schedule,
		Message:   req.Message,
		Channel:   req.Channel,
		ChatID:    req.ChatID,
		Timezone:  "Local",
		Enabled:   enabled,
		NextRun:   &now, // fire on next poll cycle; scheduler computes proper next tick after first run
		CreatedAt: now,
	}
	if err := s.cronStore.SaveCronJob(r.Context(), s.cronTenantID, rec); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "job": jobToJSON(*rec)})
}

func (s *Server) handleUpdateCronJob(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "cron store not initialised"})
		return
	}
	id := r.PathValue("id")
	if id == "" {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "id is required"})
		return
	}

	// Currently only `enabled` is mutable from the UI. Schedule edits
	// are delete + recreate to keep the audit trail simple.
	var req struct {
		Enabled *bool `json:"enabled,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "invalid request"})
		return
	}
	rec, err := s.cronStore.GetCronJob(r.Context(), s.cronTenantID, id)
	if err != nil || rec == nil {
		jsonResponse(w, http.StatusNotFound,
			map[string]any{"ok": false, "error": "job not found"})
		return
	}
	if req.Enabled != nil {
		rec.Enabled = *req.Enabled
		// When re-enabling a previously paused job, give it a fresh
		// NextRun so the scheduler picks it up on the next tick
		// instead of immediately re-firing all the missed slots.
		if *req.Enabled {
			next := time.Now()
			rec.NextRun = &next
		}
	}
	if err := s.cronStore.SaveCronJob(r.Context(), s.cronTenantID, rec); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true, "job": jobToJSON(*rec)})
}

func (s *Server) handleDeleteCronJob(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "cron store not initialised"})
		return
	}
	id := r.PathValue("id")
	if id == "" {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "id is required"})
		return
	}
	if err := s.cronStore.DeleteCronJob(r.Context(), s.cronTenantID, id); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{"ok": true})
}

// handleRunCronJobNow forces a job's NextRun to "now" so the scheduler
// will fire it on the next poll cycle (within ~60s). Useful for smoke-
// testing newly-created jobs without waiting for the schedule.
//
// Note: we don't fire synchronously here. That would require the setup
// server to know about the message bus, and would also bypass the
// scheduler's distributed lock — both of which are worse than waiting
// up to a minute.
func (s *Server) handleRunCronJobNow(w http.ResponseWriter, r *http.Request) {
	if s.cronStore == nil {
		jsonResponse(w, http.StatusServiceUnavailable,
			map[string]any{"ok": false, "error": "cron store not initialised"})
		return
	}
	id := r.PathValue("id")
	if id == "" {
		jsonResponse(w, http.StatusBadRequest,
			map[string]any{"ok": false, "error": "id is required"})
		return
	}
	rec, err := s.cronStore.GetCronJob(r.Context(), s.cronTenantID, id)
	if err != nil || rec == nil {
		jsonResponse(w, http.StatusNotFound,
			map[string]any{"ok": false, "error": "job not found"})
		return
	}
	now := time.Now()
	rec.NextRun = &now
	if err := s.cronStore.SaveCronJob(r.Context(), s.cronTenantID, rec); err != nil {
		jsonResponse(w, http.StatusInternalServerError,
			map[string]any{"ok": false, "error": err.Error()})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]any{
		"ok":      true,
		"message": "job will fire on the next scheduler poll (within 60s)",
	})
}

// newCronJobID returns a v4-style UUID. Duplicated from
// internal/agent/tools and internal/gateway because all three layers
// can independently insert jobs and we want zero cross-package
// dependencies just for ID generation.
func newCronJobID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("ts-%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

