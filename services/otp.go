package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
)

// GENERATE A UNIFORM 6 DIGIT CODE USING A CSPRNG
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HMAC THE CODE WITH A SERVER SECRET AND THE USER ID, SO A LEAKED DB CANNOT BE BRUTE FORCED OFFLINE
func HashOTP(secret string, userID uint, code string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	fmt.Fprintf(mac, "%d:%s", userID, code)
	return hex.EncodeToString(mac.Sum(nil))
}
