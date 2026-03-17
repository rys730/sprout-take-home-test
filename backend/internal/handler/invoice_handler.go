package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"sprout-backend/internal/domain"
	"sprout-backend/internal/infrastructure/logger"

	"github.com/labstack/echo/v4"
)

type InvoiceHandler struct {
	repo domain.InvoiceRepository
}

func NewInvoiceHandler(repo domain.InvoiceRepository) *InvoiceHandler {
	return &InvoiceHandler{repo: repo}
}

// ListInvoices godoc
// @Summary     List invoices
// @Description Retrieve a paginated list of invoices with optional customer_id and status filters
// @Tags        invoices
// @Produce     json
// @Param       customer_id query string false "Filter by customer UUID"
// @Param       status      query string false "Comma-separated status filter (unpaid, partially_paid, paid)"
// @Param       limit       query int    false "Page size (default 20)"
// @Param       offset      query int    false "Offset (default 0)"
// @Success     200 {object} map[string]interface{}
// @Failure     500 {object} map[string]string
// @Router      /invoices [get]
func (h *InvoiceHandler) ListInvoices(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	var statuses []string
	if raw := c.QueryParam("status"); raw != "" {
		for _, s := range strings.Split(raw, ",") {
			if t := strings.TrimSpace(s); t != "" {
				statuses = append(statuses, t)
			}
		}
	}

	filter := domain.InvoiceFilter{
		CustomerID: c.QueryParam("customer_id"),
		Statuses:   statuses,
		Limit:      limit,
		Offset:     offset,
	}

	invoices, total, err := h.repo.List(c.Request().Context(), filter)
	if err != nil {
		logger.Errorf("Failed to list invoices: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve invoices",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  invoices,
		"count": len(invoices),
		"total": total,
	})
}

// GetInvoice godoc
// @Summary     Get invoice by ID
// @Description Retrieve a single invoice by UUID
// @Tags        invoices
// @Produce     json
// @Param       id path string true "Invoice UUID"
// @Success     200 {object} map[string]interface{}
// @Failure     404 {object} map[string]string
// @Router      /invoices/{id} [get]
func (h *InvoiceHandler) GetInvoice(c echo.Context) error {
	id := c.Param("id")

	invoice, err := h.repo.GetByID(c.Request().Context(), id)
	if err != nil || invoice == nil {
		logger.Errorf("Failed to get invoice %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Invoice not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": invoice,
	})
}

// CreateInvoice godoc
// @Summary     Create invoice
// @Description Create a new customer invoice (auto-generates invoice number)
// @Tags        invoices
// @Accept      json
// @Produce     json
// @Param       body body domain.CreateInvoiceRequest true "Invoice payload"
// @Success     201 {object} map[string]interface{}
// @Failure     400 {object} map[string]string
// @Failure     500 {object} map[string]string
// @Router      /invoices [post]
func (h *InvoiceHandler) CreateInvoice(c echo.Context) error {
	ctx := c.Request().Context()

	var req domain.CreateInvoiceRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	logger.Infof("CreateInvoice request: customer_id=%s issue_date=%s due_date=%s amount=%f",
		req.CustomerID, req.IssueDate, req.DueDate, req.TotalAmount)

	if req.CustomerID == "" || req.IssueDate == "" || req.DueDate == "" || req.TotalAmount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "customer_id, issue_date, due_date, and total_amount are required",
		})
	}

	const dateFmt = "2006-01-02"
	if _, err := time.Parse(dateFmt, req.IssueDate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "issue_date must be in YYYY-MM-DD format",
		})
	}
	if _, err := time.Parse(dateFmt, req.DueDate); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "due_date must be in YYYY-MM-DD format",
		})
	}

	// Auto-generate invoice number
	invoiceNumber, err := h.repo.GenerateInvoiceNumber(ctx)
	if err != nil {
		logger.Errorf("Failed to generate invoice number: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate invoice number",
		})
	}

	// Build the domain invoice
	createdBy := "system"
	invoice := &domain.Invoice{
		InvoiceNumber: invoiceNumber,
		CustomerID:    req.CustomerID,
		IssueDate:     req.IssueDate,
		DueDate:       req.DueDate,
		TotalAmount:   req.TotalAmount,
		Description:   req.Description,
		CreatedBy:     &createdBy,
	}

	created, err := h.repo.Create(ctx, invoice)
	if err != nil {
		logger.Errorf("Failed to create invoice: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create invoice",
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data": created,
	})
}
