package service

import "errors"

var ErrEmailTaken = errors.New("email is already registered")
var ErrUserNotVerified = errors.New("user is not verified")
var ErrExpiredRefreshToken = errors.New("expired refresh token")
var ErrInvalidOldPassword = errors.New("invalid old password")
var ErrSKUTaken = errors.New("SKU product is already existed")
var ErrProductNotFound = errors.New("product not found")
var ErrCartItemNotFound = errors.New("cart item not found")
var ErrInsufficientStock = errors.New("insufficient stock")

const pgUniqueViolationCode = "23505"
