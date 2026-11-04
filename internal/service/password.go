package service

import "golang.org/x/crypto/bcrypt"

// HashPassword menghasilkan hash bcrypt untuk disimpan di users.password_hash.
func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

// CheckPassword membandingkan password input dengan hash tersimpan.
func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
