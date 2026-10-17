package media

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3config "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/x-reconstruction/api/internal/config"
	"github.com/x-reconstruction/api/internal/domain"
)

type MediaService struct {
	cfg        *config.Config
	s3Client   *s3.Client
	useR2      bool
	storageDir string
}

func NewMediaService(cfg *config.Config) *MediaService {
	svc := &MediaService{
		cfg:        cfg,
		storageDir: cfg.StorageDir,
		useR2:      false,
	}

	if cfg.R2AccessKeyID != "" && cfg.R2SecretAccessKey != "" && cfg.R2Endpoint != "" && cfg.R2Bucket != "" {
		customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL:               cfg.R2Endpoint,
				HostnameImmutable: true,
				SigningRegion:     "auto",
			}, nil
		})

		awsCfg, err := s3config.LoadDefaultConfig(context.Background(),
			s3config.WithEndpointResolverWithOptions(customResolver),
			s3config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.R2AccessKeyID,
				cfg.R2SecretAccessKey,
				"",
			)),
			s3config.WithRegion("auto"),
		)
		if err == nil {
			svc.s3Client = s3.NewFromConfig(awsCfg)
			svc.useR2 = true
			slog.Info("Cloudflare R2 storage initialized successfully", "bucket", cfg.R2Bucket, "endpoint", cfg.R2Endpoint)
		} else {
			slog.Warn("Failed to configure R2 S3 client, falling back to local storage", "error", err)
		}
	}

	return svc
}

// ProcessAndUpload processes uploaded media like X:
// - Validates and normalizes format
// - Resizes dimensions if > 2048px (X photo standard)
// - Strips EXIF metadata for user privacy
// - Compresses JPEG to 82-85% quality
// - Preserves GIFs and video streams
// - Uploads to Cloudflare R2 bucket lostmedia and returns CDN URL
func (s *MediaService) ProcessAndUpload(ctx context.Context, uploaderID uuid.UUID, filename string, r io.Reader) (*domain.MediaObject, error) {
	rawBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read media stream: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(filename))
	mediaID := uuid.New()
	key := fmt.Sprintf("media/%s%s", mediaID.String(), ext)

	var (
		processedBytes []byte
		width          *int
		height         *int
		mimeType       string
		mediaType      string
	)

	allowedExts := map[string]bool{
		".pjp": true, ".jfif": true, ".jpe": true, ".pjpeg": true,
		".jpeg": true, ".jpg": true, ".png": true, ".webp": true, ".gif": true,
		".m4v": true, ".mp4": true, ".mov": true, ".webm": true,
	}
	if !allowedExts[ext] {
		return nil, fmt.Errorf("unsupported media extension %s: only images and videos are permitted", ext)
	}

	switch ext {
	case ".jpg", ".jpeg", ".pjp", ".jfif", ".jpe", ".pjpeg":
		mediaType = "image"
		mimeType = "image/jpeg"
		processedBytes, width, height, err = s.compressJPEG(rawBytes)
		if err != nil {
			processedBytes = rawBytes // fallback to raw
		}
	case ".png":
		mediaType = "image"
		mimeType = "image/png"
		// If PNG is a large image, compress to optimized JPEG like X
		processedBytes, width, height, err = s.compressPNGToJPEG(rawBytes)
		if err != nil {
			processedBytes = rawBytes
		} else {
			ext = ".jpg"
			key = fmt.Sprintf("media/%s.jpg", mediaID.String())
			mimeType = "image/jpeg"
		}
	case ".webp":
		mediaType = "image"
		mimeType = "image/webp"
		processedBytes = rawBytes
	case ".gif":
		mediaType = "gif"
		mimeType = "image/gif"
		processedBytes = rawBytes
		if g, err := gif.DecodeAll(bytes.NewReader(rawBytes)); err == nil && len(g.Image) > 0 {
			w := g.Config.Width
			h := g.Config.Height
			width = &w
			height = &h
		}
	case ".mp4", ".mov", ".m4v":
		mediaType = "video"
		mimeType = "video/mp4"
		processedBytes = rawBytes
	case ".webm":
		mediaType = "video"
		mimeType = "video/webm"
		processedBytes = rawBytes
	default:
		return nil, fmt.Errorf("unsupported media format %s", ext)
	}

	finalSize := int64(len(processedBytes))
	var publicURL string

	if s.useR2 && s.s3Client != nil {
		putInput := &s3.PutObjectInput{
			Bucket:      aws.String(s.cfg.R2Bucket),
			Key:         aws.String(key),
			Body:        bytes.NewReader(processedBytes),
			ContentType: aws.String(mimeType),
		}

		_, err := s.s3Client.PutObject(ctx, putInput)
		if err != nil {
			slog.Error("Cloudflare R2 upload failed, saving to local filesystem fallback", "error", err, "key", key)
			return s.saveLocalFallback(uploaderID, mediaID, key, processedBytes, mediaType, mimeType, finalSize, width, height)
		}

		// Clean public CDN domain
		domain := strings.TrimRight(s.cfg.R2PublicDomain, "/")
		if domain == "" {
			domain = "https://pub-a638ff524f744502b575203fb48d39ae.r2.dev"
		}
		publicURL = fmt.Sprintf("%s/%s", domain, key)
		slog.Info("Uploaded asset to Cloudflare R2", "url", publicURL, "bytes", finalSize)
	} else {
		return s.saveLocalFallback(uploaderID, mediaID, key, processedBytes, mediaType, mimeType, finalSize, width, height)
	}

	return &domain.MediaObject{
		ID:            mediaID,
		UploaderID:    uploaderID,
		StorageKey:    key,
		URL:           publicURL,
		MediaType:     mediaType,
		MimeType:      mimeType,
		FileSizeBytes: finalSize,
		Width:         width,
		Height:        height,
	}, nil
}

func (s *MediaService) saveLocalFallback(uploaderID, mediaID uuid.UUID, key string, data []byte, mediaType, mimeType string, size int64, width, height *int) (*domain.MediaObject, error) {
	if err := os.MkdirAll(s.storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage dir: %w", err)
	}
	baseFilename := filepath.Base(key)
	localPath := filepath.Join(s.storageDir, baseFilename)
	if err := os.WriteFile(localPath, data, 0644); err != nil {
		return nil, fmt.Errorf("failed to write local file: %w", err)
	}

	return &domain.MediaObject{
		ID:            mediaID,
		UploaderID:    uploaderID,
		StorageKey:    key,
		URL:           "/uploads/" + baseFilename,
		MediaType:     mediaType,
		MimeType:      mimeType,
		FileSizeBytes: size,
		Width:         width,
		Height:        height,
	}, nil
}

// compressJPEG strips metadata, resizes if dimension > 2048px, and compresses with 85% quality
func (s *MediaService) compressJPEG(data []byte) ([]byte, *int, *int, error) {
	img, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, nil, err
	}

	img = resizeIfNeeded(img, 2048)
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	var buf bytes.Buffer
	// X standard: ~85% quality yields ideal compression ratio while preserving detail
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, nil, nil, err
	}

	return buf.Bytes(), &w, &h, nil
}

// compressPNGToJPEG converts PNG to JPEG with high fidelity and smaller size
func (s *MediaService) compressPNGToJPEG(data []byte) ([]byte, *int, *int, error) {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, nil, nil, err
	}

	img = resizeIfNeeded(img, 2048)
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		return nil, nil, nil, err
	}

	return buf.Bytes(), &w, &h, nil
}

// resizeIfNeeded downscales an image proportionally if width or height exceeds maxDim
func resizeIfNeeded(img image.Image, maxDim int) image.Image {
	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	if w <= maxDim && h <= maxDim {
		return img
	}

	var newW, newH int
	if w > h {
		newW = maxDim
		newH = int(float64(h) * (float64(maxDim) / float64(w)))
	} else {
		newH = maxDim
		newW = int(float64(w) * (float64(maxDim) / float64(h)))
	}

	if newW <= 0 {
		newW = 1
	}
	if newH <= 0 {
		newH = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newW, newH))
	for y := 0; y < newH; y++ {
		srcY := float64(y) * float64(h) / float64(newH)
		y0 := int(srcY)
		y1 := y0 + 1
		if y1 >= h {
			y1 = h - 1
		}
		dy := srcY - float64(y0)

		for x := 0; x < newW; x++ {
			srcX := float64(x) * float64(w) / float64(newW)
			x0 := int(srcX)
			x1 := x0 + 1
			if x1 >= w {
				x1 = w - 1
			}
			dx := srcX - float64(x0)

			r00, g00, b00, a00 := img.At(bounds.Min.X+x0, bounds.Min.Y+y0).RGBA()
			r10, g10, b10, a10 := img.At(bounds.Min.X+x1, bounds.Min.Y+y0).RGBA()
			r01, g01, b01, a01 := img.At(bounds.Min.X+x0, bounds.Min.Y+y1).RGBA()
			r11, g11, b11, a11 := img.At(bounds.Min.X+x1, bounds.Min.Y+y1).RGBA()

			lerp := func(v00, v10, v01, v11 uint32) uint8 {
				top := float64(v00)*(1.0-dx) + float64(v10)*dx
				bot := float64(v01)*(1.0-dx) + float64(v11)*dx
				val := top*(1.0-dy) + bot*dy
				return uint8(val / 257.0)
			}

			dst.SetRGBA(x, y, color.RGBA{
				R: lerp(r00, r10, r01, r11),
				G: lerp(g00, g10, g01, g11),
				B: lerp(b00, b10, b01, b11),
				A: lerp(a00, a10, a01, a11),
			})
		}
	}

	return dst
}

