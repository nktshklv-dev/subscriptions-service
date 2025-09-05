package subscriptions

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"subscriptions-service/internal/httpx"
)

type Handler struct {
	Repo   *Repository
	Logger *slog.Logger
}

func NewHandler(repo *Repository, logger *slog.Logger) *Handler {
	return &Handler{Repo: repo, Logger: logger}
}

func (h *Handler) RegisterMux(mux *http.ServeMux) {
	mux.HandleFunc("/subscriptions", h.collection)
	mux.HandleFunc("/subscriptions/", h.byID)
	mux.HandleFunc("/subscriptions/summary", h.summary)
}

func (h *Handler) summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	q := r.URL.Query()
	dto := SummaryDTO{
		From: q.Get("from"),
		To:   q.Get("to"),
	}
	if s := q.Get("user_id"); s != "" {
		dto.UserID = &s
	}
	if s := q.Get("service_name"); s != "" {
		dto.ServiceName = &s
	}
	from, to, uidPtr, svcPtr, err := dto.Parse()
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, err.Error())
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	total, err := h.Repo.Summary(ctx, from, to, uidPtr, svcPtr)
	if err != nil {
		h.Logger.Error("summary failed", "err", err)
		httpx.Error(w, http.StatusInternalServerError, "failed to get summary")
		return
	}
	httpx.JSON(w, http.StatusOK, SummaryResult{Total: total})
}

func (h *Handler) collection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		var in CreateDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		userID, service, price, start, end, err := in.Validate()
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		sub, err := h.Repo.Create(ctx, userID, service, price, start, end)
		if err != nil {
			h.Logger.Error("create subscription failed", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to create subscription")
			return
		}
		httpx.JSON(w, http.StatusCreated, sub)

	case http.MethodGet:
		q := r.URL.Query()
		var uidPtr *uuid.UUID
		if s := q.Get("user_id"); s != "" {
			uid, err := uuid.Parse(s)
			if err != nil {
				httpx.Error(w, http.StatusBadRequest, "user_id must be uuid")
				return
			}
			uidPtr = &uid
		}
		var servicePtr *string
		if s := q.Get("service_name"); s != "" {
			servicePtr = &s
		}
		limit := 20
		if s := q.Get("limit"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v > 0 && v <= 1000 {
				limit = v
			} else {
				httpx.Error(w, http.StatusBadRequest, "limit must be 1..1000")
				return
			}
		}
		offset := 0
		if s := q.Get("offset"); s != "" {
			if v, err := strconv.Atoi(s); err == nil && v >= 0 {
				offset = v
			} else {
				httpx.Error(w, http.StatusBadRequest, "offset must be >= 0")
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		list, err := h.Repo.List(ctx, uidPtr, servicePtr, limit, offset)
		if err != nil {
			h.Logger.Error("list subscriptions failed", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to list subscriptions")
			return
		}
		httpx.JSON(w, http.StatusOK, list)

	default:
		w.Header().Set("Allow", http.MethodPost+", "+http.MethodGet)
		httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *Handler) byID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/subscriptions/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		httpx.Error(w, http.StatusBadRequest, "missing id")
		return
	}
	id, err := uuid.Parse(parts[0])
	if err != nil {
		httpx.Error(w, http.StatusBadRequest, "id must be uuid")
		return
	}

	switch r.Method {
	case http.MethodGet:
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		sub, err := h.Repo.GetByID(ctx, id)
		if err == ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			h.Logger.Error("get subscription failed", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to get subscription")
			return
		}
		httpx.JSON(w, http.StatusOK, sub)

	case http.MethodDelete:
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		err := h.Repo.DeleteByID(ctx, id)
		if err == ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			h.Logger.Error("delete subscription failed", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to delete subscription")
			return
		}
		w.WriteHeader(http.StatusNoContent)

	case http.MethodPut:
		var in UpdateDTO
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			httpx.Error(w, http.StatusBadRequest, "invalid JSON body")
			return
		}
		userID, service, price, start, end, err := in.Validate()
		if err != nil {
			httpx.Error(w, http.StatusBadRequest, err.Error())
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		sub, err := h.Repo.UpdateByID(ctx, id, userID, service, price, start, end)
		if err == ErrNotFound {
			httpx.Error(w, http.StatusNotFound, "not found")
			return
		}
		if err != nil {
			h.Logger.Error("update subscription failed", "err", err)
			httpx.Error(w, http.StatusInternalServerError, "failed to update subscription")
			return
		}
		httpx.JSON(w, http.StatusOK, sub)

	default:
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodDelete+", "+http.MethodPut)
		httpx.Error(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
