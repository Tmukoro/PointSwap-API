package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"postswapapi/utils"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	cloudinary *cloudinary.Cloudinary
}

func NewUploadHandler() (*UploadHandler, error) {
	cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
	apiKey := os.Getenv("CLOUDINARY_API_KEY")
	apiSecret := os.Getenv("CLOUDINARY_API_SECRET")

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("cloudinary credentials not configured")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cloudinary: %w", err)
	}

	return &UploadHandler{
		cloudinary: cld,
	}, nil
}

// UploadImage uploads one or multiple images to Cloudinary
// POST /api/upload/image?type=profile (single) or type=product (multiple)
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Get user ID from context (authentication check)
	_, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get image type from query parameter
	imageType := c.DefaultQuery("type", "general")

	// Determine folder based on type
	var folder string
	var transformation string
	var maxFiles int

	switch imageType {
	case "profile":
		folder = "pointswap/profiles"
		transformation = "c_fill,g_face,h_400,w_400/q_auto,f_auto"
		maxFiles = 1 // Profile: single image only
	case "product":
		folder = "pointswap/products"
		transformation = "q_auto,f_auto"
		maxFiles = 5 // Products: up to 5 images
	case "chat":
		folder = "pointswap/chat"
		transformation = "q_auto,f_auto"
		maxFiles = 1 // Chat: single image only
	default:
		folder = "pointswap/general"
		transformation = "q_auto,f_auto"
		maxFiles = 1
	}

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid form data")
		return
	}

	files := form.File["images"] // Note: "images" (plural) for multiple files

	if len(files) == 0 {
		files = form.File["image"] // Fallback to singular
	}

	if len(files) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "no image files provided")
		return
	}

	// Check if exceeds max files
	if len(files) > maxFiles {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("too many files, max %d allowed", maxFiles))
		return
	}

	// Upload all files
	var uploadedUrls []string
	var publicIds []string

	for _, file := range files {
		// Validate file size (max 5MB per file)
		if file.Size > 5*1024*1024 {
			utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("file %s too large, max 5MB", file.Filename))
			return
		}

		// Open the uploaded file
		fileContent, err := file.Open()
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to read image")
			return
		}
		defer fileContent.Close()

		// Upload to Cloudinary
		uploadResult, err := h.cloudinary.Upload.Upload(
			context.Background(),
			fileContent,
			uploader.UploadParams{
				Folder:         folder,
				ResourceType:   "image",
				Transformation: transformation,
			},
		)

		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to upload image")
			return
		}

		uploadedUrls = append(uploadedUrls, uploadResult.SecureURL)
		publicIds = append(publicIds, uploadResult.PublicID)
	}

	// Return response
	utils.SuccessResponse(c, http.StatusOK, "images uploaded successfully", gin.H{
		"image_urls": uploadedUrls,
		"public_ids": publicIds,
		"count":      len(uploadedUrls),
		"type":       imageType,
	})
}
