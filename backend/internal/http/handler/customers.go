package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/heqiucheng/qiling-agent/backend/internal/httpx"
	"github.com/heqiucheng/qiling-agent/backend/internal/service"
)

type CustomersHandler struct {
	Service *service.QilingService
}

func (h CustomersHandler) List(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJSON(w, r, http.StatusOK, h.Service.Customers(r, httpx.ActorFromRequest(r)))
}

func (h CustomersHandler) Detail(w http.ResponseWriter, r *http.Request) {
	customerID := strings.TrimPrefix(r.URL.Path, "/api/customers/")
	detail, err := h.Service.CustomerDetail(customerID, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, detail)
}

func (h CustomersHandler) ShortTermMemory(w http.ResponseWriter, r *http.Request) {
	customerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/customers/"), "/short-term-memory")
	result, err := h.Service.CustomerShortTermMemory(customerID, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h CustomersHandler) LongTermMemory(w http.ResponseWriter, r *http.Request) {
	customerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/customers/"), "/long-term-memory")
	result, err := h.Service.CustomerLongTermMemory(customerID, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h CustomersHandler) RejectMemoryFact(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/customers/")
	parts := strings.Split(path, "/")
	if len(parts) != 5 || parts[1] != "long-term-memory" || parts[2] != "facts" || parts[4] != "reject" {
		http.NotFound(w, r)
		return
	}

	var req service.RejectMemoryFactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		WriteServiceError(w, r, err)
		return
	}
	result, err := h.Service.RejectMemoryFact(parts[0], parts[3], req, httpx.ActorFromRequest(r), httpx.RequestIDFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}

func (h CustomersHandler) Conversations(w http.ResponseWriter, r *http.Request) {
	customerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/customers/"), "/conversations")
	result, err := h.Service.CustomerConversations(customerID, r, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}
