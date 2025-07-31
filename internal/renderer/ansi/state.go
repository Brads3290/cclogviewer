package ansi

// ANSIState tracks the current state of ANSI formatting
type ANSIState struct {
	Bold          bool
	Italic        bool
	Underline     bool
	StrikeThrough bool
	FgColor       string
	BgColor       string
}

// NewANSIState creates a new ANSI state with default values
func NewANSIState() *ANSIState {
	return &ANSIState{}
}

// ApplyCodes applies ANSI codes to update the current state
func (s *ANSIState) ApplyCodes(codes []int, colorMapper *ColorMapper) {
	i := 0
	for i < len(codes) {
		code := codes[i]

		switch code {
		case 0: // Reset all
			s.Reset()

		// Text attributes
		case 1: // Bold/Bright
			s.Bold = true
		case 3: // Italic
			s.Italic = true
		case 4: // Underline
			s.Underline = true
		case 9: // Strike through
			s.StrikeThrough = true

		// Reset text attributes
		case 22: // Normal intensity (not bold)
			s.Bold = false
		case 23: // Not italic
			s.Italic = false
		case 24: // Not underlined
			s.Underline = false
		case 29: // Not crossed out
			s.StrikeThrough = false

		// Standard foreground colors (30-37)
		case 30, 31, 32, 33, 34, 35, 36, 37:
			if color, ok := colorMapper.GetForegroundColor(code); ok {
				s.FgColor = color
			}

		// Extended foreground color
		case 38:
			if i+2 < len(codes) {
				if codes[i+1] == 5 { // 256 color mode
					s.FgColor = colorMapper.Get256Color(codes[i+2])
					i += 2
				} else if codes[i+1] == 2 && i+4 < len(codes) { // RGB mode
					s.FgColor = colorMapper.GetRGBColor(codes[i+2], codes[i+3], codes[i+4])
					i += 4
				}
			}

		// Default foreground color
		case 39:
			s.FgColor = ""

		// Standard background colors (40-47)
		case 40, 41, 42, 43, 44, 45, 46, 47:
			if color, ok := colorMapper.GetBackgroundColor(code); ok {
				s.BgColor = color
			}

		// Extended background color
		case 48:
			if i+2 < len(codes) {
				if codes[i+1] == 5 { // 256 color mode
					s.BgColor = colorMapper.Get256Color(codes[i+2])
					i += 2
				} else if codes[i+1] == 2 && i+4 < len(codes) { // RGB mode
					s.BgColor = colorMapper.GetRGBColor(codes[i+2], codes[i+3], codes[i+4])
					i += 4
				}
			}

		// Default background color
		case 49:
			s.BgColor = ""

		// Bright foreground colors (90-97)
		case 90, 91, 92, 93, 94, 95, 96, 97:
			if color, ok := colorMapper.GetForegroundColor(code); ok {
				s.FgColor = color
			}

		// Bright background colors (100-107)
		case 100, 101, 102, 103, 104, 105, 106, 107:
			if color, ok := colorMapper.GetBackgroundColor(code); ok {
				s.BgColor = color
			}
		}

		i++
	}
}

// Reset resets the state to default values
func (s *ANSIState) Reset() {
	*s = ANSIState{}
}

// GetClasses returns CSS classes based on the current state
func (s *ANSIState) GetClasses() []string {
	var classes []string

	if s.Bold {
		classes = append(classes, "ansi-bold")
	}
	if s.Italic {
		classes = append(classes, "ansi-italic")
	}
	if s.Underline {
		classes = append(classes, "ansi-underline")
	}
	if s.StrikeThrough {
		classes = append(classes, "ansi-strike")
	}

	return classes
}

// GetStyles returns CSS styles based on the current state
func (s *ANSIState) GetStyles() map[string]string {
	styles := make(map[string]string)

	if s.FgColor != "" {
		styles["color"] = s.FgColor
	}
	if s.BgColor != "" {
		styles["background-color"] = s.BgColor
	}

	return styles
}

// HasFormatting returns true if the state has any formatting applied
func (s *ANSIState) HasFormatting() bool {
	return s.Bold || s.Italic || s.Underline || s.StrikeThrough ||
		s.FgColor != "" || s.BgColor != ""
}

// Clone creates a copy of the current state
func (s *ANSIState) Clone() *ANSIState {
	return &ANSIState{
		Bold:          s.Bold,
		Italic:        s.Italic,
		Underline:     s.Underline,
		StrikeThrough: s.StrikeThrough,
		FgColor:       s.FgColor,
		BgColor:       s.BgColor,
	}
}
