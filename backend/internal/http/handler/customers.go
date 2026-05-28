package handler

import (
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

func (h CustomersHandler) Conversations(w http.ResponseWriter, r *http.Request) {
	customerID := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/api/customers/"), "/conversations")
	result, err := h.Service.CustomerConversations(customerID, r, httpx.ActorFromRequest(r))
	if err != nil {
		WriteServiceError(w, r, err)
		return
	}
	httpx.WriteJSON(w, r, http.StatusOK, result)
}
