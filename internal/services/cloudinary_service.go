package services

import (
	"context"
	"dory-backend/internal/config"
	"mime/multipart"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(file multipart.File, filename string) (string, string, error) {
	ctx := context.Background()
	cld, err := cloudinary.NewFromURL(config.AppConfig.CloudinaryURL)
	if err != nil {
		return "", "", err
	}
	// Use "raw" resource type for PDFs to preserve the original file format
	// Note: We now extract text before upload, so this is mainly for storage/backup
	uploadResult, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		ResourceType: "raw",
		Folder:       "pdfs",
	})
	if err != nil {
		return "", "", err
	}

	return uploadResult.SecureURL, uploadResult.PublicID, nil
}

func DeleteFromCloudinary(publicID string) error {
	ctx := context.Background()
	cld, _ := cloudinary.NewFromURL(config.AppConfig.CloudinaryURL)

	_, err := cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	return err
}
