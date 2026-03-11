package handler

import (
	"net/http"
	"strconv"

	"sprout-backend/internal/domain"
	"sprout-backend/internal/infrastructure/logger"

	"github.com/labstack/echo/v4"
)

// ===========================================================================
// Payment Handler
// ===========================================================================

type PaymentHandler struct {
	useCase domain.PaymentUseCase
}

func NewPaymentHandler(useCase domain.PaymentUseCase) *PaymentHandler {
	return &PaymentHandler{useCase: useCase}
}

// ListPayments godoc
// @Summary     List payments
// @Description Retrieve a paginated list of payments with optional filtering by customer
// @Tags        payments
// @Produce     json
// @Security    BearerAuth
// @Param       customer_id query    string false "Filter by customer UUID"
// @Param       limit       query    int    false "Page size (default 20)"
// @Param       offset      query    int    false "Offset (default 0)"
// @Success     200         {object} map[string]interface{} "data: []Payment, count: int, total: int"
// @Failure     500         {object} map[string]string       "Failed to retrieve payments"
// @Router      /payments [get]
func (h *PaymentHandler) ListPayments(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	filter := domain.PaymentFilter{
		CustomerID: c.QueryParam("customer_id"),
		Limit:      limit,
		Offset:     offset,
	}

	payments, total, err := h.useCase.List(c.Request().Context(), filter)
	if err != nil {
		logger.Errorf("Failed to list payments: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve payments",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  payments,
		"count": len(payments),
		"total": total,
	})
}

// GetPayment godoc
// @Summary     Get payment by ID
// @Description Retrieve a single payment with its allocations by UUID
// @Tags        payments
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Payment UUID"
// @Success     200 {object} map[string]interface{} "data: Payment (with allocations)"
// @Failure     404 {object} map[string]string       "Payment not found"
// @Router      /payments/{id} [get]
func (h *PaymentHandler) GetPayment(c echo.Context) error {
	id := c.Param("id")

	payment, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Failed to get payment %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": payment,
	})
}

// RecordPayment godoc
// @Summary     Record a customer payment
// @Description Record a lump-sum payment from a customer, allocate to invoices, and auto-create journal entry (Debit Bank, Credit Piutang Usaha 112.000)
// @Tags        payments
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body     domain.CreatePaymentRequest true "Payment data with allocations"
// @Success     201  {object} map[string]interface{}       "data: Payment, message: string"
// @Failure     400  {object} map[string]string             "Validation error"
// @Router      /payments [post]
func (h *PaymentHandler) RecordPayment(c echo.Context) error {
	var req domain.CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.CustomerID == "" || req.PaymentDate == "" || req.DepositToAccountID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "customer_id, payment_date, and deposit_to_account_id are required",
		})
	}

	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "amount must be greater than zero",
		})
	}

	if len(req.Allocations) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "at least one allocation is required",
		})
	}

	createdBy := getUserIDFromContext(c)

	payment, err := h.useCase.RecordPayment(c.Request().Context(), req, createdBy)
	if err != nil {
		logger.Errorf("Failed to record payment: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data":    payment,
		"message": "Payment recorded successfully",
	})
}

// GetReceivablesSummary godoc
// @Summary     Get receivables summary
// @Description Returns total outstanding receivables (Total Piutang) and total overdue receivables (Total Jatuh Tempo)
// @Tags        payments
// @Produce     json
// @Security    BearerAuth
// @Success     200 {object} domain.ReceivablesSummary "Receivables summary"
// @Failure     500 {object} map[string]string          "Failed to retrieve receivables summary"
// @Router      /payments/summary [get]
func (h *PaymentHandler) GetReceivablesSummary(c echo.Context) error {
	summary, err := h.useCase.GetReceivablesSummary(c.Request().Context())
	if err != nil {
		logger.Errorf("Failed to get receivables summary: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve receivables summary",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": summary,
	})
}
