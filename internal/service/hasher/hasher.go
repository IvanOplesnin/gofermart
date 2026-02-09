package hasher

import (
	"crypto/sha256"
	"encoding/hex"
)

type HasherSHA256 struct{}

func NewSHA256() *HasherSHA256 {
	return &HasherSHA256{}
}

func (h *HasherSHA256) HashPassword(password string) (string, error) {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:]), nil
}

func (h *HasherSHA256) ComparePasswordHash(hashedPassword, password string) (bool, error) {
	newHash, _ := h.HashPassword(password)
	return hashedPassword == newHash, nil
}
