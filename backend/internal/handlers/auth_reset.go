package handlers

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/jamesc159/monmetrics/internal/notify"
)

const resetTokenTTL = time.Hour

// forgotPasswordResponse is returned for every well-formed request so the endpoint
// does not reveal which email addresses have accounts.
const forgotPasswordResponse = "If an account exists for that email, a reset link has been sent."

// ForgotPassword issues a single-use reset token and emails the reset link.
func (h *Handlers) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	email := normalizeEmail(req.Email)
	if !validateEmail(email) {
		h.sendError(w, "Please enter a valid email address", http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var user struct {
		ID       primitive.ObjectID `bson:"_id"`
		Email    string             `bson:"email"`
		IsActive bool               `bson:"is_active"`
	}
	err := h.db.Collection("users").FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == nil && user.IsActive {
		token, hash, err := newResetToken()
		if err != nil {
			fmt.Printf("Error generating reset token: %v\n", err)
		} else {
			expires := time.Now().UTC().Add(resetTokenTTL)
			_, err = h.db.Collection("users").UpdateOne(ctx, bson.M{"_id": user.ID}, bson.M{"$set": bson.M{
				"reset_token_hash":    hash,
				"reset_token_expires": expires,
				"updated_at":          time.Now().UTC(),
			}})
			if err != nil {
				fmt.Printf("Error storing reset token: %v\n", err)
			} else {
				link := h.config.FrontendURL + "/reset-password?token=" + url.QueryEscape(token)
				body := fmt.Sprintf("We received a request to reset your MonMetrics password.\n\n"+
					"Reset it here (valid for one hour):\n%s\n\n"+
					"If you did not request this, you can ignore this email.\n", link)
				if err := h.mailer.Send(ctx, notify.Message{To: user.Email, Subject: "Reset your MonMetrics password", Body: body}); err != nil {
					fmt.Printf("Error sending reset email: %v\n", err)
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": forgotPasswordResponse})
}

// ResetPassword sets a new password when given a valid, unexpired reset token.
func (h *Handlers) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.Token == "" {
		h.sendError(w, "Reset link is invalid or has expired", http.StatusBadRequest, nil)
		return
	}
	if err := validatePassword(req.Password); err != nil {
		h.sendError(w, err.Error(), http.StatusBadRequest, nil)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	passwordHash, err := h.hashPassword(req.Password)
	if err != nil {
		http.Error(w, "Error processing reset", http.StatusInternalServerError)
		return
	}

	// The filter checks the hash and expiry, and clears the token in the same write,
	// so a token can only be used once even if two requests race.
	now := time.Now().UTC()
	res, err := h.db.Collection("users").UpdateOne(ctx,
		bson.M{"reset_token_hash": hashResetToken(req.Token), "reset_token_expires": bson.M{"$gt": now}},
		bson.M{
			"$set":   bson.M{"password_hash": passwordHash, "updated_at": now},
			"$unset": bson.M{"reset_token_hash": "", "reset_token_expires": ""},
		},
	)
	if err != nil {
		fmt.Printf("Error resetting password: %v\n", err)
		http.Error(w, "Error resetting password", http.StatusInternalServerError)
		return
	}
	if res.MatchedCount == 0 {
		h.sendError(w, "Reset link is invalid or has expired", http.StatusBadRequest, nil)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password updated. You can now sign in."})
}

// newResetToken returns a random URL-safe token to email and its hash to store.
func newResetToken() (token, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	token = base64.RawURLEncoding.EncodeToString(buf)
	return token, hashResetToken(token), nil
}

func hashResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
