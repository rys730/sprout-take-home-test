package handler

import (
	"net/http"
	"strconv"

	"sprout-backend/internal/domain"
	"sprout-backend/internal/infrastructure/logger"

	"github.com/labstack/echo/v4"
)

type JournalHandler struct {
	useCase domain.JournalUseCase
}

func NewJournalHandler(useCase domain.JournalUseCase) *JournalHandler {
	return &JournalHandler{useCase: useCase}
}

// ListJournals godoc
// @Summary     List journal entries
// @Description Retrieve a paginated list of journal entries with optional filtering
// @Tags        journals
// @Produce     json
// @Security    BearerAuth
// @Param       status     query    string false "Filter by status (draft, posted, reversed)"
// @Param       source     query    string false "Filter by source (manual, payment, opening_balance, adjustment)"
// @Param       start_date query    string false "Filter from date (YYYY-MM-DD)"
// @Param       end_date   query    string false "Filter to date (YYYY-MM-DD)"
// @Param       limit      query    int    false "Page size (default 20)"
// @Param       offset     query    int    false "Offset (default 0)"
// @Success     200        {object} map[string]interface{} "data: []JournalEntry, count: int, total: int"
// @Failure     401        {object} map[string]string       "Unauthorized"
// @Failure     500        {object} map[string]string       "Failed to retrieve journal entries"
// @Router      /journals [get]
func (h *JournalHandler) ListJournals(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	filter := domain.JournalFilter{
		Status:    c.QueryParam("status"),
		Source:    c.QueryParam("source"),
		StartDate: c.QueryParam("start_date"),
		EndDate:   c.QueryParam("end_date"),
		Limit:     limit,
		Offset:    offset,
	}

	entries, total, err := h.useCase.List(c.Request().Context(), filter)
	if err != nil {
		logger.Errorf("Failed to list journal entries: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve journal entries",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  entries,
		"count": len(entries),
		"total": total,
	})
}

// GetJournal godoc
// @Summary     Get journal entry by ID
// @Description Retrieve a single journal entry with its lines by UUID
// @Tags        journals
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Journal Entry UUID"
// @Success     200 {object} map[string]interface{} "data: JournalEntry (with lines)"
// @Failure     401 {object} map[string]string       "Unauthorized"
// @Failure     404 {object} map[string]string       "Journal entry not found"
// @Router      /journals/{id} [get]
func (h *JournalHandler) GetJournal(c echo.Context) error {
	id := c.Param("id")

	entry, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Failed to get journal entry %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": entry,
	})
}

// CreateJournal godoc
// @Summary     Create a new journal entry
// @Description Create a new manual journal entry with balanced debit/credit lines. Status can be "draft" or "posted".
// @Tags        journals
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body     domain.CreateJournalRequest true "Journal entry data"
// @Success     201  {object} map[string]interface{}       "data: JournalEntry, message: string"
// @Failure     400  {object} map[string]string             "Validation error"
// @Failure     401  {object} map[string]string             "Unauthorized"
// @Router      /journals [post]
func (h *JournalHandler) CreateJournal(c echo.Context) error {
	var req domain.CreateJournalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Date == "" || req.Description == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "date and description are required",
		})
	}

	createdBy := getUserIDFromContext(c)

	entry, err := h.useCase.Create(c.Request().Context(), req, createdBy)
	if err != nil {
		logger.Errorf("Failed to create journal entry: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data":    entry,
		"message": "Journal entry created successfully",
	})
}

// UpdateJournal godoc
// @Summary     Update a draft journal entry
// @Description Update the date, description, or lines of a draft journal entry. Only draft entries can be edited.
// @Tags        journals
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path     string                      true "Journal Entry UUID"
// @Param       body body     domain.UpdateJournalRequest  true "Fields to update"
// @Success     200  {object} map[string]interface{}       "data: JournalEntry, message: string"
// @Failure     400  {object} map[string]string             "Validation error or not a draft"
// @Failure     401  {object} map[string]string             "Unauthorized"
// @Router      /journals/{id} [put]
func (h *JournalHandler) UpdateJournal(c echo.Context) error {
	id := c.Param("id")

	var req domain.UpdateJournalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Date == nil && req.Description == nil && len(req.Lines) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "At least one of date, description, or lines must be provided",
		})
	}

	entry, err := h.useCase.Update(c.Request().Context(), id, req)
	if err != nil {
		logger.Errorf("Failed to update journal entry %s: %v", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    entry,
		"message": "Journal entry updated successfully",
	})
}

// PostJournal godoc
// @Summary     Post a draft journal entry
// @Description Transition a draft journal entry to posted status. Entry must be balanced and have at least 2 lines.
// @Tags        journals
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Journal Entry UUID"
// @Success     200 {object} map[string]interface{} "data: JournalEntry, message: string"
// @Failure     400 {object} map[string]string       "Not a draft or unbalanced"
// @Failure     401 {object} map[string]string       "Unauthorized"
// @Router      /journals/{id}/post [post]
func (h *JournalHandler) PostJournal(c echo.Context) error {
	id := c.Param("id")

	entry, err := h.useCase.Post(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Failed to post journal entry %s: %v", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    entry,
		"message": "Journal entry posted successfully",
	})
}

// ReverseJournal godoc
// @Summary     Reverse a posted journal entry
// @Description Create a reversing entry that swaps debit/credit of the original. Original entry is marked as reversed.
// @Tags        journals
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path     string                      true "Journal Entry UUID"
// @Param       body body     domain.ReverseJournalRequest true "Reversal reason"
// @Success     201  {object} map[string]interface{}       "data: JournalEntry (reversing entry), message: string"
// @Failure     400  {object} map[string]string             "Not posted or missing reason"
// @Failure     401  {object} map[string]string             "Unauthorized"
// @Router      /journals/{id}/reverse [post]
func (h *JournalHandler) ReverseJournal(c echo.Context) error {
	id := c.Param("id")

	var req domain.ReverseJournalRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Reason == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "reason is required",
		})
	}

	createdBy := getUserIDFromContext(c)

	entry, err := h.useCase.Reverse(c.Request().Context(), id, req, createdBy)
	if err != nil {
		logger.Errorf("Failed to reverse journal entry %s: %v", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data":    entry,
		"message": "Journal entry reversed successfully",
	})
}

// DeleteJournal godoc
// @Summary     Delete a draft journal entry
// @Description Delete a journal entry. Only draft entries can be deleted.
// @Tags        journals
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Journal Entry UUID"
// @Success     200 {object} map[string]string "Journal entry deleted successfully"
// @Failure     400 {object} map[string]string "Not a draft"
// @Failure     401 {object} map[string]string "Unauthorized"
// @Router      /journals/{id} [delete]
func (h *JournalHandler) DeleteJournal(c echo.Context) error {
	id := c.Param("id")

	if err := h.useCase.Delete(c.Request().Context(), id); err != nil {
		logger.Errorf("Failed to delete journal entry %s: %v", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Journal entry deleted successfully",
	})
}
