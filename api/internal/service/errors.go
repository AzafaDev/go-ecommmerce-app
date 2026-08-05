package service

import "errors"

var ErrEmailTaken = errors.New("email is already registered")
var ErrUserNotVerified = errors.New("user is not verified")
var ErrExpiredRefreshToken = errors.New("expired refresh token")
