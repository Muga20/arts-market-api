package services

import (
	"mime/multipart"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func ValidateImageFile(file *multipart.FileHeader) error {
	// 5MB limit
	if file.Size > 5<<20 {
		return fiber.NewError(fiber.StatusBadRequest, "Image too large, maximum size is 5MB")
	}

	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		return fiber.NewError(fiber.StatusBadRequest, "Only image files are allowed")
	}

	return nil
}
