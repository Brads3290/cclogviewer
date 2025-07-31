package ansi

// ColorMapper handles ANSI color code to CSS color mapping
type ColorMapper struct {
	foregroundColors map[int]string
	backgroundColors map[int]string
	rgb256Palette    []string
}

// NewColorMapper creates a new color mapper with standard ANSI colors
func NewColorMapper() *ColorMapper {
	return &ColorMapper{
		foregroundColors: initForegroundColors(),
		backgroundColors: initBackgroundColors(),
		rgb256Palette:    init256ColorPalette(),
	}
}

func initForegroundColors() map[int]string {
	return map[int]string{
		// Standard colors (30-37)
		30: "#000000", // Black
		31: "#cc0000", // Red
		32: "#4e9a06", // Green
		33: "#c4a000", // Yellow
		34: "#3465a4", // Blue
		35: "#75507b", // Magenta
		36: "#06989a", // Cyan
		37: "#d3d7cf", // White

		// Bright colors (90-97)
		90: "#555753", // Bright Black
		91: "#ef2929", // Bright Red
		92: "#8ae234", // Bright Green
		93: "#fce94f", // Bright Yellow
		94: "#729fcf", // Bright Blue
		95: "#ad7fa8", // Bright Magenta
		96: "#34e2e2", // Bright Cyan
		97: "#eeeeec", // Bright White
	}
}

func initBackgroundColors() map[int]string {
	return map[int]string{
		// Standard background colors (40-47)
		40: "#000000", // Black
		41: "#cc0000", // Red
		42: "#4e9a06", // Green
		43: "#c4a000", // Yellow
		44: "#3465a4", // Blue
		45: "#75507b", // Magenta
		46: "#06989a", // Cyan
		47: "#d3d7cf", // White

		// Bright background colors (100-107)
		100: "#555753", // Bright Black
		101: "#ef2929", // Bright Red
		102: "#8ae234", // Bright Green
		103: "#fce94f", // Bright Yellow
		104: "#729fcf", // Bright Blue
		105: "#ad7fa8", // Bright Magenta
		106: "#34e2e2", // Bright Cyan
		107: "#eeeeec", // Bright White
	}
}

func init256ColorPalette() []string {
	// Initialize 256 color palette
	palette := make([]string, 256)

	// 0-15: Standard colors
	standardColors := []string{
		"#000000", "#800000", "#008000", "#808000",
		"#000080", "#800080", "#008080", "#c0c0c0",
		"#808080", "#ff0000", "#00ff00", "#ffff00",
		"#0000ff", "#ff00ff", "#00ffff", "#ffffff",
	}
	copy(palette[0:16], standardColors)

	// 16-231: 6x6x6 color cube
	for r := 0; r < 6; r++ {
		for g := 0; g < 6; g++ {
			for b := 0; b < 6; b++ {
				idx := 16 + r*36 + g*6 + b
				rv := 0
				if r > 0 {
					rv = 55 + 40*r
				}
				gv := 0
				if g > 0 {
					gv = 55 + 40*g
				}
				bv := 0
				if b > 0 {
					bv = 55 + 40*b
				}
				palette[idx] = rgbToHex(rv, gv, bv)
			}
		}
	}

	// 232-255: Grayscale
	for i := 0; i < 24; i++ {
		gray := 8 + i*10
		palette[232+i] = rgbToHex(gray, gray, gray)
	}

	return palette
}

func rgbToHex(r, g, b int) string {
	return "#" + toHex(r) + toHex(g) + toHex(b)
}

func toHex(n int) string {
	hex := "0123456789abcdef"
	return string([]byte{hex[n/16], hex[n%16]})
}

// GetForegroundColor returns the CSS color for a foreground ANSI code
func (m *ColorMapper) GetForegroundColor(code int) (string, bool) {
	color, exists := m.foregroundColors[code]
	return color, exists
}

// GetBackgroundColor returns the CSS color for a background ANSI code
func (m *ColorMapper) GetBackgroundColor(code int) (string, bool) {
	color, exists := m.backgroundColors[code]
	return color, exists
}

// Get256Color returns the CSS color for a 256-color palette index
func (m *ColorMapper) Get256Color(index int) string {
	if index >= 0 && index < len(m.rgb256Palette) {
		return m.rgb256Palette[index]
	}
	return "#ffffff" // Default to white
}

// GetRGBColor returns the CSS color for RGB values
func (m *ColorMapper) GetRGBColor(r, g, b int) string {
	// Clamp values to 0-255
	if r < 0 {
		r = 0
	} else if r > 255 {
		r = 255
	}
	if g < 0 {
		g = 0
	} else if g > 255 {
		g = 255
	}
	if b < 0 {
		b = 0
	} else if b > 255 {
		b = 255
	}

	return rgbToHex(r, g, b)
}
