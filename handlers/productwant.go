package handlers

import (
	"database/sql"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

//function to request what the user wants in exchange for what theyre swapping with

func CreateProductWant(ctx *gin.Context) {
	var req models.CreateProductWantRequest

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User is not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	//get productID

	productID := ctx.Param("product_id")

	if productID == "" {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Product ID is required")
		return
	}

	//verify if user owns the product

	var productOwner uuid.UUID

	err := config.DB.QueryRow(`
         SELECT seller_id FROM products WHERE product_id = $1
    `, productID).Scan(&productOwner)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Product not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	if productOwner != user.User_ID {
		utils.ErrorResponse(ctx, http.StatusForbidden, "You can only set wants for your own product")
		return
	}

	//check if wants already exist for this product

	var existWantID string

	err = config.DB.QueryRow(`
       SELECT want_id FROM product_wants WHERE product_id = $1 
    `, productID).Scan(&existWantID)

	var want models.ProductWants

	now := time.Now()

	//creating what you want in return request
	if err == sql.ErrNoRows {
		want = models.ProductWants{
			WantID:         uuid.New(),
			ProductID:      uuid.MustParse(productID),
			WantUserID:     user.User_ID,
			WantedCategory: req.WantedCategory,
			WantedSize:     req.WantedSize,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		_, err = config.DB.Exec(`
            INSERT INTO product_wants (want_id, product_id, want_user_id ,wanted_category, wanted_size, created_at, updated_at)
            VALUES($1, $2, $3, $4, $5, $6, $7) 
        `, want.WantID, want.ProductID, want.WantUserID, want.WantedCategory, want.WantedSize, want.CreatedAt,
			want.UpdatedAt)
	}
	//if there is update it
	if err == nil {
		want.WantID = uuid.MustParse(existWantID)
		want.ProductID = uuid.MustParse(productID)
		want.WantedCategory = req.WantedCategory
		want.WantedSize = req.WantedSize
		want.UpdatedAt = now

		_, err = config.DB.Exec(`
            UPDATE product_wants SET wanted_category = $2, wanted_size = $3, updated_at = $4
            WHERE want_id = $1
        `, want.WantID, want.WantedCategory, want.WantedSize, want.UpdatedAt)
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Product Want Saved Successfully", gin.H{
		"want_id": want.WantID,
	})

}

//get product wants for editing purposes

func GetProductWant(ctx *gin.Context) {
	productID := ctx.Param("product_id")

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	var want models.ProductWants

	err := config.DB.QueryRow(`
        SELECT pw.want_id, pw.product_id, pw.wanted_category, pw.wanted_size, pw.created_at, pw.updated_at
        FROM product_wants pw JOIN products p ON pw.product_id = p.product_id
        WHERE pw.product_id = $1 AND p.seller_id = $2
    `, productID, user.User_ID).Scan(&want.WantID, &want.ProductID, &want.WantedCategory, &want.WantedSize,
		&want.CreatedAt, &want.UpdatedAt)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Product Want not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Product want retrieved successfully", want)

}

//update product want

func UpdateProductWant(ctx *gin.Context) {
	var req models.UpdateProductWantRequest

	if err := ctx.ShouldBindBodyWithJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	productID := ctx.Param("product_id")

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	//verify ownership and update

	var ownerID uuid.UUID

	err := config.DB.QueryRow(`SELECT want_user_id FROM product_wants WHERE product_id = $1`, productID).Scan(&ownerID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	if ownerID != user.User_ID {
		utils.ErrorResponse(ctx, http.StatusForbidden, "Only user can modify their product want")
		return
	}

	_, err = config.DB.Exec(`
       UPDATE product_wants
	   SET wanted_category = $1, wanted_size = $2, updated_at = $3
	   WHERE product_id = $4
   `, req.WantedCategory, req.WantedSize, time.Now(), productID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "failed to update product want")
		return
	}

	//trigger notficiation for new want

	go TriggerNotificationsForNewWant(uuid.MustParse(productID), user.User_ID, req.WantedCategory, req.WantedSize)

	utils.SuccessResponse(ctx, http.StatusOK, "Product want updated successfully", nil)
}
