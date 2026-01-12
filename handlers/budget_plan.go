package handlers

import (
	"errors"
	"fmt"
	"go-expense-tracker-api/middleware"
	"go-expense-tracker-api/models"
	"go-expense-tracker-api/repositories"
	"go-expense-tracker-api/utils"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type BudgetPlanHandler struct {
	budgetPlanRepo *repositories.BudgetPlanRepository
	userRepo       *repositories.UserRepository
	expenseRepo    *repositories.ExpenseRepository
	validator      *validator.Validate
}

func NewBudgetPlanHandler(budgetPlanRepo *repositories.BudgetPlanRepository, userRepo *repositories.UserRepository, expenseRepo *repositories.ExpenseRepository) *BudgetPlanHandler {
	return &BudgetPlanHandler{
		budgetPlanRepo: budgetPlanRepo,
		userRepo:       userRepo,
		expenseRepo:    expenseRepo,
		validator:      validator.New(),
	}
}

// GET BUDGET PLAN TEMPLATES
// GetBudgetPlanTemplates godoc
// @Summary Get budget plan templates
// @Description Get default budget plan templates for the authenticated user
// @Tags budget-plans
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[[]models.BudgetPlan]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /budget-plans/templates [get]
func (h *BudgetPlanHandler) GetBudgetPlanTemplates(c *gin.Context) {
	templates, err := h.budgetPlanRepo.GetBudgetPlanTemplates()
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success get budget plan templates", templates)
}

// GET BUDGET PLANS BY USER ID
// GetBudgetPlanByUserID godoc
// @Summary Get budget plans by user ID
// @Description Get budget plans for the authenticated user with pagination and filters
// @Tags budget-plans
// @Accept  json
// @Produce  json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of items per page" default(10)
// @Param sortBy query string false "Sort by field" default(id)
// @Param order query string false "Sort order (asc or desc)" default(asc)
// @Success 200 {object} utils.ResponseWithPagination[[]models.BudgetPlan]
// @Failure 401 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /budget-plans [get]
func (h *BudgetPlanHandler) GetBudgetPlans(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// VALIDATE USER ID
	if err := h.validator.Var(userID, "required,number"); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
		return
	}

	// GET QUERY PARAMS
	queryParams, _ := c.Get("queryParams")

	budgetPlans, total, totalPages, err := h.budgetPlanRepo.GetBudgetPlansByUserID(userID.(uint), queryParams.(middleware.QueryParams))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get budget plans")
		return
	}

	response := gin.H{
		"data":        budgetPlans,
		"total":       total,
		"page":        queryParams.(middleware.QueryParams).Page,
		"limit":       queryParams.(middleware.QueryParams).Limit,
		"total_pages": totalPages,
	}

	utils.SuccessResponse(c, http.StatusOK, "Success get budget plans", response)
}

// CreateBudgetPlan godoc
// @Summary Create a new budget plan
// @Description Create a new budget plan for the authenticated user. Can be created from a template or from scratch.
// @Tags budget-plans
// @Accept json
// @Produce json
// @Param budgetPlan body models.BudgetPlanRequest true "Budget Plan"
// @Success 201 {object} utils.Response[models.BudgetPlan]
// @Failure 400 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /budget-plans [post]
// @Security BearerAuth
func (h *BudgetPlanHandler) CreateBudgetPlan(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// VALIDATE USER ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	var req models.BudgetPlanRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var totalPercentage float64

	// CALCULATE TOTAL PERCENTAGE
	for _, bucket := range req.Buckets {
		totalPercentage += bucket.Percentage
	}

	// VALIDATE TOTAL PERCENTAGE
	if totalPercentage != 100 {
		utils.ErrorResponse(c, http.StatusBadRequest, "Total percentage must be 100%")
		return
	}

	// CONVERT BUCKET REQUESTS TO BUCKETS
	var buckets []models.BudgetBucket
	for _, bucketReq := range req.Buckets {
		buckets = append(buckets, models.BudgetBucket{
			Name:         bucketReq.Name,
			Percentage:   bucketReq.Percentage,
			BucketTypeID: bucketReq.BucketTypeID,
		})
	}

	// CREATE BUDGET PLAN
	budgetPlan := &models.BudgetPlan{
		Name:        req.Name,
		UserID:      &user.ID,
		IsTemplate:  false,
		Description: req.Description,
		Buckets:     buckets,
	}

	if err := h.budgetPlanRepo.CreateBudgetPlan(budgetPlan); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create budget plan.")
		return
	}

	// SET AS ACTIVE IF SPECIFIED
	if req.SetAsActive {
		user.ActiveBudgetPlanID = &budgetPlan.ID
		if err := h.userRepo.Update(user); err != nil {
			log.Printf("Failed to set budget plan %d as active for user %d: %v", budgetPlan.ID, user.ID, err)
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, "Budget plan created successfully", budgetPlan)
}

// GetBudgetPlanByID godoc
// @Summary Get budget plan by ID
// @Description Get a budget plan by its ID for the authenticated user
// @Tags budget-plans
// @Accept json
// @Produce json
// @Param id path int true "Budget Plan ID"
// @Success 200 {object} utils.Response[models.BudgetPlan]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /budget-plans/{id} [get]
// @Security BearerAuth
func (h *BudgetPlanHandler) GetBudgetPlan(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// VALIDATE USER ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// GET BUDGET PLAN ID FROM URL PARAMS
	budgetPlanID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid budget plan ID")
		return
	}

	// GET BUDGET PLAN FROM REPOSITORY
	budgetPlan, err := h.budgetPlanRepo.GetBudgetPlanByID(uint(budgetPlanID), user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Budget plan not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get budget plan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Success get budget plan", budgetPlan)
}

// UpdateBudgetPlan godoc
// @Summary Update a budget plan
// @Description Update an existing budget plan for the authenticated user
// @Tags budget-plans
// @Accept json
// @Produce json
// @Param id path int true "Budget Plan ID"
// @Param budgetPlan body models.BudgetPlanRequest true "Budget Plan"
// @Success 200 {object} utils.Response[models.BudgetPlan]
// @Failure 400 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /budget-plans/{id} [put]
// @Security BearerAuth
func (h *BudgetPlanHandler) UpdateBudgetPlan(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// VALIDATE USER ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// GET BUDGET PLAN ID FROM URL PARAMS
	budgetPlanID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid budget plan ID")
		return
	}

	// GET BUDGET PLAN FROM REPOSITORY
	budgetPlan, err := h.budgetPlanRepo.GetBudgetPlanByID(uint(budgetPlanID), user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Budget plan not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get budget plan")
		return
	}

	var req models.BudgetPlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// INPUT VALIDATION
	if err := h.validator.Struct(req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var totalPercentage float64
	bucketsChanged := len(req.Buckets) > 0
	budgetPlan.Name = req.Name
	budgetPlan.Description = req.Description

	if bucketsChanged {
		var bucketsToCreate []models.BudgetBucket
		var bucketsToUpdate []models.BudgetBucket

		existingBucketsMap := make(map[uint]models.BudgetBucket)
		for _, b := range budgetPlan.Buckets {
			existingBucketsMap[b.ID] = b
		}

		processedBucketIDs := make(map[uint]struct{})

		for _, bucketReq := range req.Buckets {
			totalPercentage += bucketReq.Percentage

			// Process update bucket
			if bucketReq.ID != nil {
				if _, ok := existingBucketsMap[*bucketReq.ID]; !ok {
					utils.ErrorResponse(c, http.StatusBadRequest, "Invalid bucket ID for update: "+strconv.Itoa(int(*bucketReq.ID)))
					return
				}
				bucketsToUpdate = append(bucketsToUpdate, models.BudgetBucket{
					ID:           *bucketReq.ID,
					BudgetPlanID: budgetPlan.ID,
					Name:         bucketReq.Name,
					Percentage:   bucketReq.Percentage,
				})
				processedBucketIDs[*bucketReq.ID] = struct{}{}
				// Process create bucket
			} else {
				bucketsToCreate = append(bucketsToCreate, models.BudgetBucket{
					BudgetPlanID: budgetPlan.ID,
					Name:         bucketReq.Name,
					Percentage:   bucketReq.Percentage,
				})
			}
		}

		if totalPercentage != 100 && len(req.Buckets) > 0 {
			utils.ErrorResponse(c, http.StatusBadRequest, "Total percentage must be 100%")
			return
		}

		var bucketIDsToDelete []uint
		for _, existingBucket := range budgetPlan.Buckets {
			if _, ok := processedBucketIDs[existingBucket.ID]; !ok {
				bucketIDsToDelete = append(bucketIDsToDelete, existingBucket.ID)
			}
		}

		if err := h.budgetPlanRepo.UpdateBudgetPlanAndSyncBuckets(budgetPlan, bucketsToCreate, bucketsToUpdate, bucketIDsToDelete); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update budget plan.")
			return
		}
	} else {
		if err := h.budgetPlanRepo.UpdateBudgetPlanAndSyncBuckets(budgetPlan, nil, nil, nil); err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update budget plan.")
			return
		}
	}

	// RE-FETCH THE BUDGET PLAN TO GET THE LATEST BUCKET DATA
	updatedBudgetPlan, err := h.budgetPlanRepo.GetBudgetPlanByID(uint(budgetPlanID), user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve updated budget plan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Budget plan updated successfully", updatedBudgetPlan)
}

// DeleteBudgetPlan godoc
// @Summary Delete a budget plan
// @Description Delete an existing budget plan for the authenticated user
// @Tags budget-plans
// @Produce json
// @Param id path int true "Budget Plan ID"
// @Success 200 {object} utils.Response[models.DeleteBudgetPlanResponse]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 403 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Router /budget-plans/{id} [delete]
// @Security BearerAuth
func (h *BudgetPlanHandler) DeleteBudgetPlan(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// VALIDATE USER ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// GET BUDGET PLAN ID FROM URL PARAMS
	budgetPlanID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid budget plan ID")
		return
	}

	// GET BUDGET PLAN FROM REPOSITORY
	budgetPlan, err := h.budgetPlanRepo.GetBudgetPlanByID(uint(budgetPlanID), user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Budget plan not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get budget plan")
		return
	}

	// CHECK IF BUDGET PLAN IS BELONGING TO USER
	if budgetPlan.UserID == nil || *budgetPlan.UserID != user.ID {
		utils.ErrorResponse(c, http.StatusForbidden, "You do not have permission to delete this budget plan")
		return
	}

	if err := h.budgetPlanRepo.DeleteBudgetPlan(budgetPlan); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete budget plan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Budget plan deleted successfully.", budgetPlan)
}

// SetActiveBudgetPlan godoc
// @Summary Set an active budget plan
// @Description Set a budget plan as active for the authenticated user
// @Tags budget-plans
// @Accept  json
// @Produce  json
// @Param id path int true "Budget Plan ID"
// @Success 200 {object} utils.Response[models.User]
// @Failure 400 {object} utils.Response[any]
// @Failure 401 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /budget-plans/{id}/set-active [post]
func (h *BudgetPlanHandler) SetActiveBudgetPlan(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// VALIDATE USER ID
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// GET BUDGET PLAN ID FROM URL PARAMS
	budgetPlanID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid budget plan ID")
		return
	}

	// GET BUDGET PLAN FROM REPOSITORY
	budgetPlan, err := h.budgetPlanRepo.GetBudgetPlanByID(uint(budgetPlanID), user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Budget plan not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get budget plan")
		return
	}

	user.ActiveBudgetPlanID = &budgetPlan.ID
	if err := h.userRepo.Update(user); err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to set active budget plan")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Budget plan has been set as active", user)
}

// GET ACTIVE BUDGET PLAN
// GetActiveBudgetPlan godoc
// @Summary Get active budget plan
// @Description Get the active budget plan for the authenticated user, along with total spending for each bucket type and their percentage of income
// @Tags budget-plans
// @Accept  json
// @Produce  json
// @Success 200 {object} utils.Response[models.ActiveBudgetPlanResponse]
// @Failure 401 {object} utils.Response[any]
// @Failure 404 {object} utils.Response[any]
// @Failure 500 {object} utils.Response[any]
// @Security BearerAuth
// @Router /budget-plans/active [get]
func (h *BudgetPlanHandler) GetActiveBudgetPlan(c *gin.Context) {
	// GET USER ID FROM CONTEXT
	userID, exists := c.Get("user_id")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// GET USER
	user, err := h.userRepo.GetByID(userID.(uint))
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid user ID")
		return
	}

	// CHECK IF USER HAS ACTIVE BUDGET PLAN
	if user.ActiveBudgetPlanID == nil {
		utils.ErrorResponse(c, http.StatusNotFound, "No active budget plan found")
		return
	}

	// GET ACTIVE BUDGET PLAN
	budgetPlan, err := h.budgetPlanRepo.GetBudgetPlanByID(*user.ActiveBudgetPlanID, user.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ErrorResponse(c, http.StatusNotFound, "Active budget plan not found")
			return
		}
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get active budget plan")
		return
	}

	// GET TOTAL INCOME
	totalIncome, totalExpenses, err := h.expenseRepo.GetTotalIncomeAndExpenses(user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get total income")
		return
	}
	fmt.Printf("Total Income: %.2f, Total Expenses: %.2f\n", totalIncome, totalExpenses)

	// GET TOTAL SPENDING FOR EACH BUCKET
	spendingByBucket, err := h.budgetPlanRepo.GetTotalSpendingByBucket(*user.ActiveBudgetPlanID, user.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get spending by bucket")
		return
	}

	// CREATE MAP FOR SPENDING
	spendingMap := make(map[string]float64)
	for _, s := range spendingByBucket {
		spendingMap[s.BucketName] = s.TotalSpending
	}

	// CALCULATE ALLOCATION AND PERCENTAGE
	var bucketResponses []models.BudgetBucketResponse
	for _, bucket := range budgetPlan.Buckets {
		maxAllocation := (bucket.Percentage / 100) * totalIncome
		totalSpending := spendingMap[bucket.Name]
		spendingPercentage := 0.0
		if maxAllocation > 0 {
			spendingPercentage = (totalSpending / maxAllocation) * 100
		}

		bucketResponses = append(bucketResponses, models.BudgetBucketResponse{
			BudgetBucket:       bucket,
			MaxAllocation:      maxAllocation,
			TotalSpending:      totalSpending,
			SpendingPercentage: spendingPercentage,
		})
	}

	response := models.ActiveBudgetPlanResponse{
		BudgetPlan:  *budgetPlan,
		TotalIncome: totalIncome,
		Buckets:     bucketResponses,
	}

	utils.SuccessResponse(c, http.StatusOK, "Success get active budget plan", response)
}
