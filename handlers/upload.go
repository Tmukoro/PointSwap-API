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

// UploadImage uploads an image to Cloudinary and returns the URL
// POST /api/upload/image
func (h *UploadHandler) UploadImage(c *gin.Context) {
	// Get user ID from context (authentication check)
	_, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Get the uploaded file
	file, err := c.FormFile("image")
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "no image file provided")
		return
	}

	// Validate file size (max 5MB)
	if file.Size > 5*1024*1024 {
		utils.ErrorResponse(c, http.StatusBadRequest, "image too large, max 5MB")
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
			Folder:         "pointswap/chat", // Organize in folder
			ResourceType:   "image",
			Transformation: "q_auto,f_auto", // Auto quality and format
		},
	)

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "failed to upload image")
		return
	}

	// Return the secure URL
	utils.SuccessResponse(c, http.StatusOK, "image uploaded successfully", gin.H{
		"image_url": uploadResult.SecureURL,
		"public_id": uploadResult.PublicID,
	})
}
