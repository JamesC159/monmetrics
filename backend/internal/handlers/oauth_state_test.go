package handlers

import (
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestOAuthStateRoundTrip(t *testing.T) {
	secret := []byte("test-secret")
	user := primitive.NewObjectID()
	now := time.Now()

	state, err := signOAuthState(secret, user, "ebay", now)
	if err != nil {
		t.Fatal(err)
	}
	got, err := verifyOAuthState(secret, state, "ebay", now)
	if err != nil || got != user {
		t.Fatalf("verify failed: %v", err)
	}

	if _, err := verifyOAuthState(secret, state, "tcgplayer", now); err == nil {
		t.Fatal("expected provider mismatch error")
	}
	if _, err := verifyOAuthState(secret, state, "ebay", now.Add(oauthStateTTL+time.Second)); err == nil {
		t.Fatal("expected expiry error")
	}
	if _, err := verifyOAuthState([]byte("other"), state, "ebay", now); err == nil {
		t.Fatal("expected signature error")
	}
	tampered := strings.Replace(state, state[:4], "AAAA", 1)
	if _, err := verifyOAuthState(secret, tampered, "ebay", now); err == nil {
		t.Fatal("expected tamper error")
	}
}
