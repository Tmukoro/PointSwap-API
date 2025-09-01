package handlers

import (
	"database/sql"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// @Summary Creates the product to swap for upload
// @Description Handle uploading of a product into the app/feed
func CreateProduct(ctx *gin.Context) {
	var req models.CreateProductRequest

	if err := ctx.ShouldBind(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	//Checks if user is authorised or basically if theyre signed in

	presentUser, exist := ctx.Get("User")
	if !exist {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not Authenticated")
		return
	}

	user, ok := presentUser.(models.Users)
	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Invalid User")
		return
	}

	//Creating the product for swapping

	product := models.Products{
		Product_ID:     uuid.New(),
		Seller_ID:      user.User_ID,
		Title:          req.Title,
		Category:       req.Category,
		Estimated_size: &req.Estimated_size,
		Status:         "active",
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
	}

	//Process for product and photos

	tx, err := config.DB.Begin()
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to start transavtion")
		return
	}

	defer tx.Rollback()

	//Putting in the products to the db

	_, err = tx.Exec(`
	   INSERT INTO products (product_id, seller_id, title, category, estimated_size, status, created_at, updated_at)
	   VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, product.Product_ID, product.Seller_ID, product.Title, product.Category, product.Estimated_size,
		product.Status, product.Created_at, product.Updated_at,
	)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create product")
		return
	}

	//Putting in phots to the db

	var photos []models.ProductPhotos

	for i, imageUrl := range req.Image_Urls {
		photo := models.ProductPhotos{
			Photo_ID:      uuid.New(),
			Product_ID:    product.Product_ID,
			Image_Url:     imageUrl,
			Display_order: i + 1,
			Created_at:    time.Now(),
		}

		_, err = tx.Exec(`
	    INSERT INTO product_photos (photo_id, product_id, image_url, display_order, created_at)
		VALUES($1, $2, $3, $4, $5)
	 `, photo.Photo_ID, photo.Product_ID, photo.Image_Url, photo.Display_order, photo.Created_at)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save photos")
			return
		}

		photos = append(photos, photo)

	}

	//commit changes

	if err = tx.Commit(); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to commit changes")
		return
	}

	//Preparing the response with photos

	response := models.ProductWithSeller{
		Product_ID:     product.Product_ID,
		Seller:         user,
		Title:          product.Title,
		Category:       product.Category,
		Estimated_size: *product.Estimated_size,
		Status:         product.Status,
		Photos:         photos,
		Created_at:     time.Now(),
		Updated_at:     time.Now(),
	}

	utils.SuccessResponse(ctx, http.StatusCreated, "Product Successfully Created", gin.H{
		"product":       response,
		"ask_for_wants": true,
		"product_id":    product.Product_ID,
	})
}

//function to get products uploaded by users in the main page

func GetProducts(ctx *gin.Context) {
	//parse pagination parameters

	limitStr := ctx.DefaultQuery("limit", "20")
	offsetStr := ctx.DefaultQuery("offset", "0")
	category := ctx.Query("category")

	limit, err := strconv.Atoi(limitStr)

	if err != nil || limit <= 0 || limit > 100 {
		limit = 20
	}

	offset, err := strconv.Atoi(offsetStr)

	if err != nil || offset < 0 {
		offset = 0
	}

	presentUser, exist := ctx.Get("User")

	var currentUserID uuid.UUID

	if exist {
		if user, ok := presentUser.(models.Users); ok {
			currentUserID = user.User_ID
		}
	}

	//Query for what to see in the main feed

	var query string
	var args []any
	argIndex := 1

	query = `
	   SELECT p.product_id, p.title, p.estimated_size, p.created_at, pp.image_url
	   FROM products p LEFT JOIN product_photos pp ON p.product_id = pp.product_id AND pp.display_order = 1
	   WHERE p.status = $1
	`
	args = append(args, "active")
	argIndex++

	//exclude current signed in users product from the feed

	if currentUserID != uuid.Nil {
		query += "AND p.seller_id != $" + strconv.Itoa(argIndex)
		args = append(args, currentUserID)
		argIndex++
	}

	//Category filter in the feed when applied

	if category != "" {
		query += " AND p.category = $" + strconv.Itoa(argIndex)
		args = append(args, category)
		argIndex++
	}

	query += " ORDER BY p.created_at DESC LIMIT $" + strconv.Itoa(argIndex)
	args = append(args, limit+1) // this is to check if there are more items
	argIndex++

	query += " OFFSET $" + strconv.Itoa(argIndex)
	args = append(args, offset)
	argIndex++

	rows, err := config.DB.Query(query, args...)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	defer rows.Close()

	//Simple structure for the feed

	var products []gin.H
	for rows.Next() {
		var productId uuid.UUID
		var title string
		var estimatedSize *string
		var createdAt time.Time
		var imageUrl sql.NullString

		err := rows.Scan(&productId, &title, &estimatedSize, &createdAt, &imageUrl)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to parse products")
			return
		}

		product := gin.H{
			"product_id":     productId,
			"title":          title,
			"estimated_size": estimatedSize,
			"created_at":     createdAt,
			"image_url":      imageUrl,
		}

		if imageUrl.Valid {
			product["image_url"] = imageUrl.String
		}

		products = append(products, product)
	}

	//Checking if there are more items when scrolling
	hasMore := len(products) > limit
	if hasMore {
		products = products[:limit]
	}

	//Calculating next offset

	var nextOffset *int
	if hasMore {
		next := offset + limit
		nextOffset = &next
	}

	response := models.InfiniteScrollData{
		Items: products,
		Meta: models.PaginationMeta{
			Limit:       limit,
			Offset:      offset,
			Has_more:    hasMore,
			Next_offset: nextOffset,
		},
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Products retrieved successfully", response)

}

//Get product via ID (Basically when you tap on the product)

func GetProductById(ctx *gin.Context) {
	productIDStr := ctx.Param("product_id")
	productID, err := uuid.Parse(productIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid ID")
		return
	}

	//incase of FE activity reconsider adding the seller details to be scanned in the query

	var product models.ProductWithSeller

	err = config.DB.QueryRow(`
	    SELECT p.product_id, p.seller_id, p.title, p.category, p.estimated_size,
		p.status, p.created_at, p.updated_at
		FROM products p JOIN users u ON p.seller_id = user_id
		WHERE p.product_id = $1
	`, productID).Scan(&product.Product_ID, &product.Seller.User_ID, &product.Title, &product.Category, &product.Estimated_size,
		&product.Status, &product.Created_at, &product.Updated_at)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Product not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch product")
		return
	}

	//Get all photos in the product view

	photoRows, err := config.DB.Query(`
	 SELECT photo_id, image_url, display_order, created_at
	 FROM product_photos
	 WHERE product_id = $1 ORDER BY display_order
	`, productID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch photos")
		return
	}

	defer photoRows.Close()

	var photos []models.ProductPhotos

	for photoRows.Next() {
		var photo models.ProductPhotos
		photo.Product_ID = productID

		err := photoRows.Scan(&photo.Photo_ID, &photo.Image_Url, &photo.Display_order, &photo.Created_at)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, err.Error())
			return
		}

		photos = append(photos, photo)
	}

	product.Photos = photos

	utils.SuccessResponse(ctx, http.StatusOK, "Product retrieved successfully", product)
}

//Viewing all the users app listed on the app(for the user themself)

func GetMyProducts(ctx *gin.Context) {
	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "User not valid")
		return
	}

	rows, err := config.DB.Query(`
	  SELECT product_id, title, category, estimated_size, status, created_at, updated_at
	  FROM products 
	  WHERE seller_id = $1
	  ORDER BY created_at DESC
	`, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch your products")
		return
	}

	defer rows.Close()

	var products []models.Products

	for rows.Next() {
		var product models.Products

		product.Seller_ID = user.User_ID

		err := rows.Scan(&product.Product_ID, &product.Title, &product.Category, &product.Estimated_size,
			&product.Status, &product.Created_at, &product.Updated_at,
		)

		if err != nil {
			utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to parse prodcut")
			return
		}

		products = append(products, product)
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved products", products)

}

//Function to update product status

func UpdateProductStatus(ctx *gin.Context) {
	productIDstr := ctx.Param("product_id")
	productID, err := uuid.Parse(productIDstr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	var req struct {
		Status string `json:"status" binding:"required,oneof=active swapped inactive" `
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	presentUser, exists := ctx.Get("User")
	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)
	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "User not valid")
		return
	}

	var ownerID uuid.UUID

	err = config.DB.QueryRow("SELECT seller_id FROM products WHERE product_id = $1 ", productID).Scan(&ownerID)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Product Not found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database error")
		return
	}

	if ownerID != user.User_ID {
		utils.ErrorResponse(ctx, http.StatusForbidden, "You can only update your own products!")
		return
	}

	//updating the status of the product

	_, err = config.DB.Exec(`
	  UPDATE products
	  SET status = $1, updated_at = $2
	  WHERE product_id = $3
	`, req.Status, time.Now(), productID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update product status")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Successfully Updated Product Status", gin.H{
		"product_id": productID,
		"status":     req.Status,
	})
}

//Removing product (set the product status as inactive hence a soft delete)

func DeleteProduct(ctx *gin.Context) {
	productIDStr := ctx.Param("product_id")
	productID, err := uuid.Parse(productIDStr)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, "Invalid Product ID")
		return
	}

	presentUser, exists := ctx.Get("User")

	if !exists {
		utils.ErrorResponse(ctx, http.StatusUnauthorized, "User not authenticated")
		return
	}

	user, ok := presentUser.(models.Users)

	if !ok {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "User not valid")
		return
	}

	//verify if product belongs to user
	var ownerID uuid.UUID

	err = config.DB.QueryRow("SELECT seller_id FROM products WHERE product_id = $1", productID).Scan(&ownerID)

	if err == sql.ErrNoRows {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Product Not Found")
		return
	}

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Database Error")
		return
	}

	if ownerID != user.User_ID {
		utils.ErrorResponse(ctx, http.StatusForbidden, "Only User can delete their product")
		return
	}

	//Start transaction for hard delete

	tx, err := config.DB.Begin()

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to start transaction")
		return
	}

	defer tx.Rollback()

	//Delete product photos first

	_, err = tx.Exec(`DELETE FROM product_photos WHERE product_id = $1`, productID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete photos")
		return
	}

	//Delete product wants

	_, err = tx.Exec(`DELETE FROM product_wants WHERE product_id = $1`, productID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete product wants")
		return
	}

	//Delete the product entirely

	result, err := tx.Exec(`DELETE FROM products WHERE product_id = $1`, productID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete product")
		return
	}

	//check if row was actually deleted

	rowsAffected, err := result.RowsAffected()

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to verify deletion")
		return
	}

	if rowsAffected == 0 {
		utils.ErrorResponse(ctx, http.StatusNotFound, "Product not found")
		return
	}

	//commit transaction

	if err = tx.Commit(); err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to commit deletion")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Product Successfully deleted", gin.H{
		"product_id": productID,
	})

}
