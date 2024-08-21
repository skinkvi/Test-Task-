package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
	"github.com/skinkvi/TestTaskBackDev/internal/config"
	database "github.com/skinkvi/TestTaskBackDev/internal/db"
	"github.com/skinkvi/TestTaskBackDev/internal/models"
)

func GenerateTokenPair(w http.ResponseWriter, r *http.Request, cfg config.Config) {
	userID := chi.URLParam(r, "userID")
	ipAddress := r.RemoteAddr

	logrus.Infof("Generating token pair for user: %s, ip: %s", userID, ipAddress)

	accessToken, err := GenerateAccessToken(userID, ipAddress, cfg)
	if err != nil {
		logrus.Errorf("Failed to generate access token for user: %s, ip: %s, err: %v", userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Generated access token: %s", accessToken)

	refreshToken, err := GenerateRefreshToken(userID, ipAddress)
	if err != nil {
		logrus.Errorf("Failed to generate refresh token for user: %s, ip: %s, err: %v", userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Generated refresh token: %s", refreshToken)

	hashedRefreshToken, err := HashRefreshToken(refreshToken)
	if err != nil {
		logrus.Errorf("Failed to hash refresh token for user: %s, ip: %s, err: %v", userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Hashed refresh token: %s", hashedRefreshToken)

	// прилошлось делать такую махинацию потому что выдавало ошибку с типом int не понял почему она была
	var userIDInt int64
	userIDInt, err = strconv.ParseInt(userID, 10, 64)
	if err != nil {
		logrus.Errorf("Failed to parse userID: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	refreshTokenRecord := models.RefreshToken{
		UserID:    int(userIDInt),
		TokenHash: hashedRefreshToken,
		ExpiresAt: time.Now().Add(time.Hour * 24 * 7),
		IPAddress: ipAddress,
	}
	err = database.DB.Create(&refreshTokenRecord).Error
	if err != nil {
		logrus.Errorf("Failed to create refresh token record for user: %s, ip: %s, err: %v", userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("Created refresh token record for user: %s, ip: %s", userID, ipAddress)

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("Failed to serialize response json for user: %s, ip: %s, err: %v", userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
