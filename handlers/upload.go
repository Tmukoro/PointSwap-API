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
	_, err := getUserIDFromContext(c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
		return
	}

	imageType := c.DefaultQuery("type", "general")

	var folder string
	var transformation string
	var maxFiles int
	var resourceType string // Add this

	switch imageType {
	case "profile":
		folder = "pointswap/profiles"
		transformation = "c_fill,g_face,h_400,w_400/q_auto,f_auto"
		maxFiles = 1
		resourceType = "image"
	case "product":
		folder = "pointswap/products"
		transformation = "q_auto,f_auto"
		maxFiles = 5
		resourceType = "image"
	case "chat":
		folder = "pointswap/chat"
		transformation = "q_auto,f_auto"
		maxFiles = 1
		resourceType = "image"
	case "audio":
		folder = "pointswap/audio"
		transformation = ""
		maxFiles = 1
		resourceType = "video" // Cloudinary uses "video" for audio files
	default:
		folder = "pointswap/general"
		transformation = "q_auto,f_auto"
		maxFiles = 1
		resourceType = "image"
	}

	form, err := c.MultipartForm()
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "invalid form data")
		return
	}

	files := form.File["images"]
	if len(files) == 0 {
		files = form.File["image"]
	}
	if len(files) == 0 {
		files = form.File["audio"] // Add support for audio field
	}

	if len(files) == 0 {
		utils.ErrorResponse(c, http.StatusBadRequest, "no files provided")
		return
	}

	if len(files) > maxFiles {
		utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("too many files, max %d allowed", maxFiles))
		return
	}

	var uploadedUrls []string
	var publicIds []string

	for _, file := range files {
		if file.Size > 10*1024*1024 { // 10MB for audio
			utils.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("file %s too large, max 10MB", file.Filename))
			return
		}

		fileContent, err := file.Open()
		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to read file")
			return
		}
		defer fileContent.Close()

		uploadParams := uploader.UploadParams{
			Folder:       folder,
			ResourceType: resourceType,
		}

		if transformation != "" {
			uploadParams.Transformation = transformation
		}

		uploadResult, err := h.cloudinary.Upload.Upload(
			context.Background(),
			fileContent,
			uploadParams,
		)

		if err != nil {
			utils.ErrorResponse(c, http.StatusInternalServerError, "failed to upload file")
			return
		}

		uploadedUrls = append(uploadedUrls, uploadResult.SecureURL)
		publicIds = append(publicIds, uploadResult.PublicID)
	}

	// Return response (no duration here)
	if len(uploadedUrls) == 1 {
		utils.SuccessResponse(c, http.StatusOK, "file uploaded successfully", gin.H{
			"image_url": uploadedUrls[0],
			"public_id": publicIds[0],
			"type":      imageType,
		})
	} else {
		utils.SuccessResponse(c, http.StatusOK, "files uploaded successfully", gin.H{
			"image_urls": uploadedUrls,
			"public_ids": publicIds,
			"count":      len(uploadedUrls),
			"type":       imageType,
		})
	}

}
