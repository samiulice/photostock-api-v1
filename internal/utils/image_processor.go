package utils

import (
	"image"
	"image/color"
	"path/filepath"

	"github.com/disintegration/imaging"
)

func GenerateImageVariants(originalPath, baseName string) error {
	img, err := imaging.Open(originalPath)
	if err != nil {
		return err
	}

	// --- Generate Thumbnail ---
	thumb := imaging.Thumbnail(img, 200, 200, imaging.Lanczos)
	thumbPath := filepath.Join("./images/thumbnail", "thumb_"+baseName)
	err = imaging.Save(thumb, thumbPath)
	if err != nil {
		return err
	}

	// --- Generate Watermarked Image ---
	wmImg := imaging.Resize(img, 1200, 0, imaging.Lanczos)
	wm := imaging.OverlayCenter(wmImg, createWatermark(wmImg.Bounds().Dx(), wmImg.Bounds().Dy()), 0.25)
	wmPath := filepath.Join("./images/watermarked", "wm_"+baseName)
	err = imaging.Save(wm, wmPath)
	if err != nil {
		return err
	}

	return nil
}

func createWatermark(width, height int) *image.NRGBA {
	// Simple white semi-transparent watermark image (text-based is better)
	wm := imaging.New(width, height, color.NRGBA{255, 255, 255, 0}) // Transparent base
	// For real use: render text or logo instead
	square := imaging.New(200, 80, color.NRGBA{255, 255, 255, 128}) // white semi-transparent
	return imaging.PasteCenter(wm, square)
}
