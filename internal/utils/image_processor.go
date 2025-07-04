package utils

import (
	"fmt"
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
)

const (
	WatermarkText     = "Photostock"
	WatermarkFontSize = 48
	WatermarkOpacity  = 0.5 // overall alpha; color will be black or white
	WatermarkMaxWidth = 1200
	FontFile          = "./assets/fonts/arial.ttf"
)

// getAverageBrightness samples the image at a grid (every step pixels) and
// returns a value in [0,1], where 0 is pure black and 1 is pure white.
func getAverageBrightness(img image.Image, step int) float64 {
	bounds := img.Bounds()
	var sum float64
	var count float64
	for y := bounds.Min.Y; y < bounds.Max.Y; y += step {
		for x := bounds.Min.X; x < bounds.Max.X; x += step {
			r, g, b, _ := img.At(x, y).RGBA()
			// RGBA returns 0–65535, convert to [0,1]
			lum := (float64(r) + float64(g) + float64(b)) / 3.0 / 65535.0
			sum += lum
			count++
		}
	}
	if count == 0 {
		return 1.0
	}
	return sum / count
}

func generateWatermarked(original image.Image, outputPath string) error {
	// 1. Resize to max width
	resized := imaging.Resize(original, WatermarkMaxWidth, 0, imaging.Lanczos)

	// 2. Compute brightness
	brightness := getAverageBrightness(resized, 20)
	// if image is mostly dark (<0.5) we want a white watermark, else black
	var r, g, b float64
	if brightness < 0.5 {
		r, g, b = 1.0, 1.0, 1.0 // white
	} else {
		r, g, b = 0.0, 0.0, 0.0 // black
	}

	// 3. Create drawing context
	w := resized.Bounds().Dx()
	h := resized.Bounds().Dy()
	dc := gg.NewContext(w, h)
	dc.DrawImage(resized, 0, 0)

	// 4. Load font
	if err := dc.LoadFontFace(FontFile, WatermarkFontSize); err != nil {
		return fmt.Errorf("load font: %w", err)
	}

	// 5. Set watermark style: dynamic color + global opacity
	dc.SetRGBA(r, g, b, WatermarkOpacity)

	// 6. Calculate the step between repeated marks
	stepx := int(float64(WatermarkFontSize) * 6)
	stepy := int(float64(WatermarkFontSize) * 8)

	// 7. Rotate the context by 45° around the center
	dc.RotateAbout(gg.Radians(-45), float64(w)/2, float64(h)/2)

	// 8. Draw tiled text across a larger virtual canvas
	for y := -h; y < h*2; y += stepx {
		for x := -w; x < w*2; x += stepy {
			dc.DrawStringAnchored(WatermarkText, float64(x), float64(y), 0.5, 0.5)
		}
	}

	// 9. Save as PNG (supports transparency)
	if err := dc.SavePNG(outputPath); err != nil {
		return fmt.Errorf("saving watermarked image: %w", err)
	}
	log.Printf("Generated tiled watermarked image: %s (brightness=%.2f)", outputPath, brightness)
	return nil
}

func GenerateImageVariants(originalPath, outputBaseDir, baseName string) error {
	thumbDir := filepath.Join(outputBaseDir, "thumbnails")
	wmDir := filepath.Join(outputBaseDir, "watermarked")
	for _, d := range []string{thumbDir, wmDir} {
		if err := os.MkdirAll(d, 0750); err != nil {
			return err
		}
	}

	img, err := imaging.Open(originalPath, imaging.AutoOrientation(true))
	if err != nil {
		return err
	}

	// Generate thumbnail
	thumb := imaging.Thumbnail(img, 300, 300, imaging.Lanczos)
	thumbPath := filepath.Join(thumbDir, "thumb_"+baseName)
	if err := imaging.Save(thumb, thumbPath); err != nil {
		return err
	}

	// Generate dynamically-colored, tiled watermark
	wmPath := filepath.Join(wmDir, "wm_"+baseName)
	if err := generateWatermarked(img, wmPath); err != nil {
		return err
	}

	return nil
}
