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

// OutputFormat represents supported output formats
type OutputFormat int

const (
	FormatWEBP OutputFormat = iota
	FormatPNG
)

// OutputOptions configures the output encoding
type OutputOptions struct {
	Format  OutputFormat
	Quality float32 // For WEBP: ignored (nativewebp only supports lossless)
}

// DefaultOutputOptions returns sensible output defaults
func DefaultOutputOptions() *OutputOptions {
	return &OutputOptions{
		Format:  FormatWEBP,
		Quality: 85,
	}
}

// SaveImage saves the rendered image to a file
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

// WriteImage writes the rendered image to a writer
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

// FormatFromPath returns the output format based on file extension
func FormatFromPath(path string) OutputFormat {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png":
		return FormatPNG
	default:
		return FormatWEBP
	}
}
