package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const oauthStateTTL = 10 * time.Minute

// signOAuthState creates a tamper-proof state value binding the OAuth flow to a user.
func signOAuthState(secret []byte, userID primitive.ObjectID, provider string, now time.Time) (string, error) {
	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	payload := strings.Join([]string{
		userID.Hex(), provider, strconv.FormatInt(now.Add(oauthStateTTL).Unix(), 10), hex.EncodeToString(nonce),
	}, "|")
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + stateMAC(secret, payload), nil
}

// verifyOAuthState checks signature, provider, and expiry, returning the bound user.
func verifyOAuthState(secret []byte, state, provider string, now time.Time) (primitive.ObjectID, error) {
	parts := strings.Split(state, ".")
	if len(parts) != 2 {
		return primitive.NilObjectID, errors.New("malformed state")
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return primitive.NilObjectID, errors.New("malformed state")
	}
	payload := string(raw)
	if !hmac.Equal([]byte(stateMAC(secret, payload)), []byte(parts[1])) {
		return primitive.NilObjectID, errors.New("invalid state signature")
	}
	fields := strings.Split(payload, "|")
	if len(fields) != 4 || fields[1] != provider {
		return primitive.NilObjectID, errors.New("state does not match provider")
	}
	exp, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil || now.Unix() > exp {
		return primitive.NilObjectID, errors.New("state expired")
	}
	userID, err := primitive.ObjectIDFromHex(fields[0])
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid state user")
	}
	return userID, nil
}

func stateMAC(secret []byte, payload string) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte("marketplace-oauth-state:" + payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
