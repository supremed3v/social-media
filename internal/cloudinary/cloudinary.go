package cloudinary

import (
	"context"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

type CloudinaryService interface {
	UploadImage(ctx context.Context, imagePath string) (string, error)
	DeleteImage(ctx context.Context, publicID string) error
	UploadImageFromReader(ctx context.Context, reader io.Reader, filePath string) (string, error)
}

type Cloudinary struct {
	cld *cloudinary.Cloudinary
}

func NewCloudinary(cfg CloudinaryConfig) (*Cloudinary, error) {
	cld, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, err
	}

	return &Cloudinary{cld: cld}, nil
}

func (c *Cloudinary) UploadImage(ctx context.Context, imagePath string) (string, error) {
	uploadParams := uploader.UploadParams{}

	uploadResult, err := c.cld.Upload.Upload(ctx, imagePath, uploadParams)

	if err != nil {
		return "", err
	}

	return uploadResult.SecureURL, nil
}

func (c *Cloudinary) DeleteImage(ctx context.Context, publicID string) error {
	_, err := c.cld.Upload.Destroy(ctx, uploader.DestroyParams{
		PublicID: publicID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (c *Cloudinary) UploadImageFromReader(ctx context.Context, reader io.Reader, file string) (string, error) {
	uploadParams := uploader.UploadParams{}

	uploadResult, err := c.cld.Upload.Upload(ctx, reader, uploadParams)

	if err != nil {
		return "", err
	}

	return uploadResult.URL, nil
}
