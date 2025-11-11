package glance

import (
	"bytes"
	"encoding/base64"
	"testing"
	"time"
)

func TestAuthTokenGenerationAndVerification(t *testing.T) {
	secret, err := makeAuthSecretKey(AUTH_SECRET_KEY_LENGTH)
	if err != nil {
		t.Fatalf("Nie udało się wygenerować tajnego klucza (secret key): %v", err)
	}

	secretBytes, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		t.Fatalf("Nie udało się zdekodować tajnego klucza (secret key): %v", err)
	}

	if len(secretBytes) != AUTH_SECRET_KEY_LENGTH {
		t.Fatalf("Tajny klucz (secret key) ma nieprawidłową długość: %d bajtów", AUTH_SECRET_KEY_LENGTH)
	}

	now := time.Now()
	username := "admin"

	token, err := generateSessionToken(username, secretBytes, now)
	if err != nil {
		t.Fatalf("Nie udało się wygenerować tokena sesji: %v", err)
	}

	usernameHashBytes, shouldRegen, err := verifySessionToken(token, secretBytes, now)
	if err != nil {
		t.Fatalf("Nie udało się zweryfikować tokena sesji: %v", err)
	}

	if shouldRegen {
		t.Fatal("Token nie powinien wymagać natychmiastowej regeneracji po wygenerowaniu")
	}

	computedUsernameHash, err := computeUsernameHash(username, secretBytes)
	if err != nil {
		t.Fatalf("Nie udało się obliczyć hasha nazwy użytkownika: %v", err)
	}

	if !bytes.Equal(usernameHashBytes, computedUsernameHash) {
		t.Fatal("Hash nazwy użytkownika nie zgadza się z oczekiwaną wartością")
	}

	// Test token regeneration
	timeRightAfterRegenPeriod := now.Add(AUTH_TOKEN_VALID_PERIOD - AUTH_TOKEN_REGEN_BEFORE + 2*time.Second)
	_, shouldRegen, err = verifySessionToken(token, secretBytes, timeRightAfterRegenPeriod)
	if err != nil {
		t.Fatalf("Weryfikacja tokena nie powinna nie powieść się w okresie regeneracji, err: %v", err)
	}

	if !shouldRegen {
		t.Fatal("Token powinien być oznaczony do regeneracji")
	}

	// Test token expiration
	_, _, err = verifySessionToken(token, secretBytes, now.Add(AUTH_TOKEN_VALID_PERIOD+2*time.Second))
	if err == nil {
		t.Fatal("Oczekiwano, że weryfikacja tokena nie powiedzie się po wygaśnięciu tokena")
	}

	// Test tampered token
	decodedToken, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		t.Fatalf("Nie udało się zdekodować tokena: %v", err)
	}

	// If any of the bytes are off by 1, the token should be considered invalid
	for i := range len(decodedToken) {
		tampered := make([]byte, len(decodedToken))
		copy(tampered, decodedToken)
		tampered[i] += 1

		_, _, err = verifySessionToken(base64.StdEncoding.EncodeToString(tampered), secretBytes, now)
		if err == nil {
			t.Fatalf("Oczekiwano, że weryfikacja tokena nie powiedzie się dla zmienionego tokena na indeksie %d", i)
		}
	}
}
