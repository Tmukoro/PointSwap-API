package handlers

import (
	"fmt"
	"math"
	"net/http"
	"postswapapi/config"
	"postswapapi/models"
	"postswapapi/utils"

	"github.com/gin-gonic/gin"
)

// DetermineLocation finds the nearest camp based on coordinates
func DetermineLocation(ctx *gin.Context) {
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

	var req models.SetLocationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(ctx, http.StatusBadRequest, err.Error())
		return
	}

	// Find nearest camp
	camp, err := findNearestCamp(req.Latitude, req.Longitude)
	if err != nil {
		utils.ErrorResponse(ctx, http.StatusNotFound, "No camp found near your location")
		return
	}

	// Update user's location
	_, err = config.DB.Exec(`
		UPDATE users 
		SET location_state = $1, camp_name = $2, latitude = $3, longitude = $4, updated_at = NOW()
		WHERE user_id = $5
	`, camp.StateName, camp.CampName, req.Latitude, req.Longitude, user.User_ID)

	if err != nil {
		utils.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update location")
		return
	}

	utils.SuccessResponse(ctx, http.StatusOK, "Location set successfully", models.SetLocationResponse{
		LocationState: camp.StateName,
		CampName:      camp.CampName,
	})
}

type Camp struct {
	StateName string
	CampName  string
	Latitude  float64
	Longitude float64
	RadiusKm  int
}

func findNearestCamp(userLat, userLng float64) (*Camp, error) {
	query := `
		SELECT state_name, camp_name, latitude, longitude, radius_km
		FROM camps
	`

	rows, err := config.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var nearestCamp *Camp
	minDistance := math.MaxFloat64

	for rows.Next() {
		var camp Camp
		err := rows.Scan(&camp.StateName, &camp.CampName, &camp.Latitude, &camp.Longitude, &camp.RadiusKm)
		if err != nil {
			continue
		}

		// Calculate distance using Haversine formula
		distance := haversineDistance(userLat, userLng, camp.Latitude, camp.Longitude)

		// Check if within radius and closer than previous camps
		if distance <= float64(camp.RadiusKm) && distance < minDistance {
			minDistance = distance
			nearestCamp = &camp
		}
	}

	if nearestCamp == nil {
		return nil, fmt.Errorf("no camp found within range")
	}

	return nearestCamp, nil
}

// Haversine formula to calculate distance between two coordinates in km
func haversineDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371 // Earth radius in kilometers

	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*
			math.Sin(dLon/2)*math.Sin(dLon/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}
