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
	const op = "handlers.GenerateTokenPair"
	userID := chi.URLParam(r, "userID")
	ipAddress := r.RemoteAddr

	logrus.Infof("%s: Generating token pair for user: %s, ip: %s", op, userID, ipAddress)

	accessToken, err := GenerateAccessToken(userID, ipAddress, cfg)
	if err != nil {
		logrus.Errorf("%s: Failed to generate access token for user: %s, ip: %s, err: %v", op, userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("%s: Generated access token: %s", op, accessToken)

	refreshToken, err := GenerateRefreshToken(userID, ipAddress)
	if err != nil {
		logrus.Errorf("%s: Failed to generate refresh token for user: %s, ip: %s, err: %v", op, userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("%s: Generated refresh token: %s", op, refreshToken)

	hashedRefreshToken, err := HashRefreshToken(refreshToken)
	if err != nil {
		logrus.Errorf("%s: Failed to hash refresh token for user: %s, ip: %s, err: %v", op, userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("%s: Hashed refresh token: %s", op, hashedRefreshToken)

	// прилошлось делать такую махинацию потому что выдавало ошибку с типом int не понял почему она была
	var userIDInt int64
	userIDInt, err = strconv.ParseInt(userID, 10, 64)
	if err != nil {
		logrus.Errorf("%s: Failed to parse userID: %v", op, err)
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
		logrus.Errorf("%s: Failed to create refresh token record for user: %s, ip: %s, err: %v", op, userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	logrus.Infof("%s: Created refresh token record for user: %s, ip: %s", op, userID, ipAddress)

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logrus.Errorf("%s: Failed to serialize response json for user: %s, ip: %s, err: %v", op, userID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}
