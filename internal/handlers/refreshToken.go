package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/skinkvi/TestTaskBackDev/internal/config"
	database "github.com/skinkvi/TestTaskBackDev/internal/db"
	"github.com/skinkvi/TestTaskBackDev/internal/models"
	"github.com/skinkvi/TestTaskBackDev/internal/templates"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

func GenerateRefreshToken(userID, ipAddress string) (string, error) {
	expirationTime := time.Now().Add(time.Hour * 24 * 7)
	token := fmt.Sprintf("%s|%s|%s", userID, ipAddress, expirationTime.Format(time.RFC3339))
	encodedToken := base64.StdEncoding.EncodeToString([]byte(token))

	return encodedToken, nil
}

func HashRefreshToken(token string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("Error hashing refresh token: %v", err)
		return "", err
	}
	return string(bytes), nil
}

func RefreshTokenPair(w http.ResponseWriter, r *http.Request, cfg config.Config) {
	refreshToken := r.Header.Get("Authorization")
	ipAddress := r.RemoteAddr

	var refreshTokenRecord models.RefreshToken
	database.DB.Where("token_hash = ?", refreshToken).First(&refreshTokenRecord)

	if refreshTokenRecord.ID == 0 {
		http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
		return
	}

	if refreshTokenRecord.IPAddress != ipAddress {
		log.Warn("IP address changed for user:", refreshTokenRecord.UserID)

		user := models.User{}
		database.DB.Model(&models.User{}).Where("id = ?", refreshTokenRecord.UserID).First(&user)

		data := struct {
			OldIP  string
			NewIP  string
			Email  string
			UserID string
		}{
			OldIP:  refreshTokenRecord.IPAddress,
			NewIP:  ipAddress,
			Email:  user.Email,
			UserID: string(user.ID),
		}

		emailBody, err := templates.ExecuteTemplate("email-warning.html", data)
		if err != nil {
			log.Errorf("Failed to generate email warning: %v", err)
			return
		}

		msg := gomail.NewMessage()
		msg.SetHeader("From", "test@task.com")
		msg.SetHeader("To", user.Email)
		msg.SetHeader("Subject", "Warning: IP address changed")
		msg.SetBody("text/html", emailBody)

		d := gomail.NewDialer("smtp.gmail.com", 587, "test@task.com", "testpassword")
		if err := d.DialAndSend(msg); err != nil {
			log.Errorf("Failed to send email warning: %v", err)
		}
	}

	accessToken, err := GenerateAccessToken(string(refreshTokenRecord.UserID), ipAddress, cfg)
	if err != nil {
		log.Error("Failed to generate access token:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	newRefreshToken, err := GenerateRefreshToken(string(refreshTokenRecord.UserID), ipAddress)
	if err != nil {
		log.Error("Failed to generate refresh token:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	hashedRefreshToken, err := HashRefreshToken(newRefreshToken)
	if err != nil {
		log.Error("Failed to hash refresh token:", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	refreshTokenRecord.TokenHash = hashedRefreshToken
	refreshTokenRecord.ExpiresAt = time.Now().Add(time.Hour * 24 * 7)
	database.DB.Save(&refreshTokenRecord)

	response := map[string]string{
		"access_token":  accessToken,
		"refresh_token": newRefreshToken,
	}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Errorf("Failed to serialize response json for user: %d, ip: %s, err: %v", refreshTokenRecord.UserID, ipAddress, err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
	log.Info("Refresh token refreshed successfully")
}
