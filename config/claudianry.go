package config

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

// CloudinaryClient holds the Cloudinary instance
type CloudinaryClient struct {
	Cloudinary *cloudinary.Cloudinary
}

// NewCloudinaryClient initializes a new Cloudinary client
func NewCloudinaryClient() (*CloudinaryClient, error) {
	cloudName := Envs.CloudinaryCloudName
	apiKey := Envs.CloudinaryAPIKey
	apiSecret := Envs.CloudinaryAPISecret

	if cloudName == "" || apiKey == "" || apiSecret == "" {
		return nil, fmt.Errorf("missing Cloudinary credentials in environment variables")
	}

	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Cloudinary: %v", err)
	}

	return &CloudinaryClient{Cloudinary: cld}, nil
}

// UploadFile uploads an image to Cloudinary
func (c *CloudinaryClient) UploadFile(fileHeader *multipart.FileHeader, folder string) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	ctx := context.Background()
	uploadResult, err := c.Cloudinary.Upload.Upload(ctx, file, uploader.UploadParams{
		Folder: folder,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Cloudinary: %v", err)
	}

	return uploadResult.SecureURL, nil
}

// DeleteFile removes an image from Cloudinary
func (c *CloudinaryClient) DeleteFile(publicID string) error {
	ctx := context.Background()
	_, err := c.Cloudinary.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from Cloudinary: %v", err)
	}
	return nil
}
