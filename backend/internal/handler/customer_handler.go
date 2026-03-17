package handler

import (
	"net/http"
	"strconv"

	"sprout-backend/internal/domain"
	"sprout-backend/internal/infrastructure/logger"

	"github.com/labstack/echo/v4"
)

type CustomerHandler struct {
	useCase domain.CustomerUseCase
}

func NewCustomerHandler(useCase domain.CustomerUseCase) *CustomerHandler {
	return &CustomerHandler{useCase: useCase}
}

// ListCustomers godoc
// @Summary     List customers
// @Description Retrieve a paginated list of customers with optional search
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       search query    string false "Search by name"
// @Param       limit  query    int    false "Page size (default 20)"
// @Param       offset query    int    false "Offset (default 0)"
// @Success     200    {object} map[string]interface{} "data: []Customer, count: int, total: int"
// @Failure     500    {object} map[string]string       "Failed to retrieve customers"
// @Router      /customers [get]
func (h *CustomerHandler) ListCustomers(c echo.Context) error {
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	offset, _ := strconv.Atoi(c.QueryParam("offset"))

	filter := domain.CustomerFilter{
		Search: c.QueryParam("search"),
		Limit:  limit,
		Offset: offset,
	}

	customers, total, err := h.useCase.List(c.Request().Context(), filter)
	if err != nil {
		logger.Errorf("Failed to list customers: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve customers",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":  customers,
		"count": len(customers),
		"total": total,
	})
}

// GetCustomer godoc
// @Summary     Get customer by ID
// @Description Retrieve a single customer by UUID
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id  path     string true "Customer UUID"
// @Success     200 {object} map[string]interface{} "data: Customer"
// @Failure     404 {object} map[string]string       "Customer not found"
// @Router      /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(c echo.Context) error {
	id := c.Param("id")

	customer, err := h.useCase.GetByID(c.Request().Context(), id)
	if err != nil {
		logger.Errorf("Failed to get customer %s: %v", id, err)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": customer,
	})
}

// CreateCustomer godoc
// @Summary     Create a new customer
// @Description Create a new customer record
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       body body     domain.CreateCustomerRequest true "Customer data"
// @Success     201  {object} map[string]interface{}        "data: Customer, message: string"
// @Failure     400  {object} map[string]string              "Validation error"
// @Failure     500  {object} map[string]string              "Failed to create customer"
// @Router      /customers [post]
func (h *CustomerHandler) CreateCustomer(c echo.Context) error {
	var req domain.CreateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "name is required",
		})
	}

	createdBy := getUserIDFromContext(c)

	customer, err := h.useCase.Create(c.Request().Context(), req, createdBy)
	if err != nil {
		logger.Errorf("Failed to create customer: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"data":    customer,
		"message": "Customer created successfully",
	})
}

// UpdateCustomer godoc
// @Summary     Update a customer
// @Description Update an existing customer's details
// @Tags        customers
// @Accept      json
// @Produce     json
// @Security    BearerAuth
// @Param       id   path     string                       true "Customer UUID"
// @Param       body body     domain.UpdateCustomerRequest  true "Fields to update"
// @Success     200  {object} map[string]interface{}        "data: Customer, message: string"
// @Failure     400  {object} map[string]string              "Validation error"
// @Failure     404  {object} map[string]string              "Customer not found"
// @Router      /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(c echo.Context) error {
	id := c.Param("id")

	var req domain.UpdateCustomerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	customer, err := h.useCase.Update(c.Request().Context(), id, req)
	if err != nil {
		logger.Errorf("Failed to update customer %s: %v", id, err)
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data":    customer,
		"message": "Customer updated successfully",
	})
}

// DeleteCustomer godoc
// @Summary     Delete a customer
// @Description Delete a customer by UUID
// @Tags        customers
// @Produce     json
// @Security    BearerAuth
// @Param       id path     string true "Customer UUID"
// @Success     200 {object} map[string]string "message: string"
// @Failure     404 {object} map[string]string "Customer not found"
// @Router      /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(c echo.Context) error {
	id := c.Param("id")

	if err := h.useCase.Delete(c.Request().Context(), id); err != nil {
		logger.Errorf("Failed to delete customer %s: %v", id, err)
		if err.Error() == "customer not found" {
			return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Customer deleted successfully",
	})
}
