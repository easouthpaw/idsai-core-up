package images

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProcessAvatar_CropsAndEncodesJPEG(t *testing.T) {
	raw := mustPNG(t, 640, 480)

	out, contentType, err := ProcessAvatar(raw)
	require.NoError(t, err)
	require.Equal(t, "image/jpeg", contentType)
	require.Less(t, len(out), MaxAvatarBytes)

	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	require.NoError(t, err)
	require.Equal(t, "jpeg", format)
	require.Equal(t, 400, cfg.Width)
	require.Equal(t, 400, cfg.Height)
}

func TestProcessProjectCover_CropsAndEncodesJPEG(t *testing.T) {
	raw := mustJPEG(t, 900, 900)

	out, contentType, err := ProcessProjectCover(raw)
	require.NoError(t, err)
	require.Equal(t, "image/jpeg", contentType)

	cfg, format, err := image.DecodeConfig(bytes.NewReader(out))
	require.NoError(t, err)
	require.Equal(t, "jpeg", format)
	require.Equal(t, 1280, cfg.Width)
	require.Equal(t, 720, cfg.Height)
}

func TestProcessImageRejectsInvalidInput(t *testing.T) {
	_, _, err := ProcessAvatar(nil)
	require.ErrorIs(t, err, ErrInvalidImageData)

	_, _, err = ProcessAvatar([]byte("not an image"))
	require.ErrorIs(t, err, ErrInvalidImageType)

	_, _, err = ProcessAvatar(mustPNG(t, 120, 120))
	require.ErrorIs(t, err, ErrImageTooSmall)

	_, _, err = ProcessProjectCover([]byte{0xff, 0xd8, 0xff, 0xdb})
	require.ErrorIs(t, err, ErrInvalidImageData)
}

func mustPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := gradient(width, height)
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}

func mustJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := gradient(width, height)
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}))
	return buf.Bytes()
}

func gradient(width, height int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.SetRGBA(x, y, color.RGBA{
				R: uint8((x * 255) / max(width-1, 1)),
				G: uint8((y * 255) / max(height-1, 1)),
				B: 180,
				A: 255,
			})
		}
	}
	return img
}
