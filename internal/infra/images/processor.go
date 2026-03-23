package images

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/color"
	stddraw "image/draw"
	"image/jpeg"
	_ "image/png"
	"net/http"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

const (
	MaxAvatarBytes       = 8 << 20
	MaxProjectCoverBytes = 12 << 20
)

var (
	ErrInvalidImageType = errors.New("invalid image type")
	ErrInvalidImageData = errors.New("invalid image data")
	ErrImageTooSmall    = errors.New("image is too small")
)

func ProcessAvatar(raw []byte) ([]byte, string, error) {
	if err := validateType(raw); err != nil {
		return nil, "", err
	}
	img, cfg, err := decode(raw)
	if err != nil {
		return nil, "", err
	}
	if cfg.Width < 400 || cfg.Height < 400 {
		return nil, "", ErrImageTooSmall
	}

	cropped := cropToAspect(img, 1.0)
	resized := resize(cropped, 400, 400)
	out, err := encodeJPEG(resized, 84)
	if err != nil {
		return nil, "", err
	}
	return out, "image/jpeg", nil
}

func ProcessProjectCover(raw []byte) ([]byte, string, error) {
	if err := validateType(raw); err != nil {
		return nil, "", err
	}
	img, _, err := decode(raw)
	if err != nil {
		return nil, "", err
	}

	cropped := cropToAspect(img, 16.0/9.0)
	resized := resize(cropped, 1280, 720)
	out, err := encodeJPEG(resized, 82)
	if err != nil {
		return nil, "", err
	}
	return out, "image/jpeg", nil
}

func validateType(raw []byte) error {
	if len(raw) == 0 {
		return ErrInvalidImageData
	}
	contentType := http.DetectContentType(raw)
	switch contentType {
	case "image/jpeg", "image/png", "image/webp":
		return nil
	default:
		return ErrInvalidImageType
	}
}

func decode(raw []byte) (image.Image, image.Config, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return nil, image.Config{}, ErrInvalidImageData
	}
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, image.Config{}, ErrInvalidImageData
	}
	return img, cfg, nil
}

func cropToAspect(src image.Image, targetAspect float64) image.Image {
	b := src.Bounds()
	w := float64(b.Dx())
	h := float64(b.Dy())
	if w <= 0 || h <= 0 {
		return src
	}

	srcAspect := w / h
	crop := b
	if srcAspect > targetAspect {
		newW := int(h * targetAspect)
		if newW < 1 {
			newW = 1
		}
		left := b.Min.X + (b.Dx()-newW)/2
		crop = image.Rect(left, b.Min.Y, left+newW, b.Max.Y)
	} else if srcAspect < targetAspect {
		newH := int(w / targetAspect)
		if newH < 1 {
			newH = 1
		}
		top := b.Min.Y + (b.Dy()-newH)/2
		crop = image.Rect(b.Min.X, top, b.Max.X, top+newH)
	}

	dst := image.NewRGBA(image.Rect(0, 0, crop.Dx(), crop.Dy()))
	stddraw.Draw(dst, dst.Bounds(), src, crop.Min, stddraw.Src)
	return dst
}

func resize(src image.Image, width, height int) image.Image {
	if width <= 0 || height <= 0 {
		return src
	}
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	stddraw.Draw(dst, dst.Bounds(), &image.Uniform{C: color.White}, image.Point{}, stddraw.Src)
	xdraw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), xdraw.Over, nil)
	return dst
}

func encodeJPEG(src image.Image, quality int) ([]byte, error) {
	if quality <= 0 || quality > 100 {
		quality = 82
	}
	buf := bytes.NewBuffer(make([]byte, 0, 64*1024))
	if err := jpeg.Encode(buf, src, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("encode image: %w", err)
	}
	return buf.Bytes(), nil
}
