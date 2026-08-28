package hash

func HashPassword(password string) (string, error) {
	// bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(password), nil
}

func CheckPasswordHash(password, hash string) bool {
	// err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return  password == hash
}
