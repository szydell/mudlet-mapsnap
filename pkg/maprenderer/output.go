package maprenderer

import (
	"fmt"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/HugoSmits86/nativewebp"
)

// OutputFormat represents the supported image output formats.
type OutputFormat int

const (
	// FormatWEBP outputs lossless WEBP images (default).
	FormatWEBP OutputFormat = iota
	// FormatPNG outputs PNG images with best compression.
	FormatPNG
)

// OutputOptions configures the image encoding behavior.
type OutputOptions struct {
	// Format specifies the output image format.
	Format OutputFormat
	// Quality is reserved for future lossy WEBP support (currently unused).
	Quality float32
}

// DefaultOutputOptions returns default output options (lossless WEBP).
func DefaultOutputOptions() *OutputOptions {
	return &OutputOptions{
		Format:  FormatWEBP,
		Quality: 85,
	}
}

// SaveImage saves the rendered image to a file at the specified path.
//
// The output format is auto-detected from the file extension:
//   - .webp: Lossless WEBP format
//   - .png: PNG format with best compression
//
// Pass nil for opts to use [DefaultOutputOptions].
func SaveImage(img *image.RGBA, path string, opts *OutputOptions) error {
	if opts == nil {
		opts = DefaultOutputOptions()
	}

	// Auto-detect format from extension if not explicitly set
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".webp":
		opts.Format = FormatWEBP
	case ".png":
		opts.Format = FormatPNG
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating output file: %w", err)
	}
	defer f.Close()

	return WriteImage(img, f, opts)
}

// WriteImage writes the rendered image to the given io.Writer.
// Pass nil for opts to use [DefaultOutputOptions].
func WriteImage(img *image.RGBA, w io.Writer, opts *OutputOptions) error {
	if opts == nil {
		opts = DefaultOutputOptions()
	}

	switch opts.Format {
	case FormatWEBP:
		return encodeWEBP(img, w)
	case FormatPNG:
		return encodePNG(img, w)
	default:
		return fmt.Errorf("unsupported output format: %d", opts.Format)
	}
}

// encodeWEBP encodes the image as lossless WEBP using nativewebp (pure Go)
func encodeWEBP(img *image.RGBA, w io.Writer) error {
	return nativewebp.Encode(w, img, nil)
}

// encodePNG encodes the image as PNG
func encodePNG(img *image.RGBA, w io.Writer) error {
	encoder := &png.Encoder{
		CompressionLevel: png.BestCompression,
	}
	return encoder.Encode(w, img)
}

// FormatFromPath determines the output format from a file path's extension.
// Returns [FormatPNG] for .png files, [FormatWEBP] for all others.
func FormatFromPath(path string) OutputFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return FormatPNG
	default:
		return FormatWEBP
	}
}
