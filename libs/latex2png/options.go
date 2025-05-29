package latex2png

import (
	"fmt"
	"image/color"
)

type OutputFormat = string

const (
	PNG OutputFormat = "--png"
	GIF              = "--gif"
)

type Options struct {
	// The paths to the used binaries
	LatexBinary  string
	DvipngBinary string

	PreambleFilePath string

	// The path to the directory to put the temporary files
	// Default : 'os.TempDir()' if empty
	TempDir string
	// Whether to add the '\begin{document}' and end to the latex input
	// Default : true
	AddBeginDocument bool
	// Default : PNG
	OutputFormat OutputFormat
	// If alpha is not 0, the color will not be transparent
	BackgroundColor color.Color
	ForegroundColor color.Color
	// Default : 100
	ImageDPI int
}

func formatColor(c color.Color) string {
	r, g, b, a := c.RGBA()

	if a == 0 {
		return "Transparent"
	}
	return fmt.Sprintf("RGB %d %d %d", r/256, g/256, b/256)
}
