package cloudinary

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func InitCloudinary() (*cloudinary.Cloudinary, context.Context, error) {
	cld, err := cloudinary.NewFromParams(os.Getenv("CLOUDINARY_NAME"), os.Getenv("CLOUDINARY_API_KEY"), os.Getenv("CLOUDINARY_SECRET_KEY"))
	if err != nil {
		return nil, nil, fmt.Errorf("Cloudinary config error: %w",err)
	}
	cld.Config.URL.Secure = true
	ctx := context.Background()
	return cld, ctx, nil
}

// Uploads any local file or multipart content and returns the secure URL
func UploadProfileImage(cld *cloudinary.Cloudinary, ctx context.Context, file interface{}, publicID string) (string, error) {
	resp, err := cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:       publicID,
		UniqueFilename: api.Bool(true),
		Overwrite:      api.Bool(false),
	})
	if err != nil {
		return "", err
	}
	return resp.SecureURL, nil
}
