package utils

type contextKey string

// LoginKey - special key, that contains user login.
const LoginKey contextKey = "userLogin"

// IDKey - special key, that containts user ID.
const IDKey contextKey = "userID"
