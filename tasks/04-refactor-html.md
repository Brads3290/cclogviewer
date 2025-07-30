# Task 04: Refactor html.go

**File:** `internal/renderer/html.go`  
**Current Size:** 389 lines  
**Priority:** MEDIUM  
**Estimated Effort:** 3-4 hours  

## Problem Summary

The `html.go` file contains mixed responsibilities:
- ANSI escape sequence parsing and conversion (150+ lines)
- HTML generation and template functions
- Complex color mapping logic with extensive switch statements
- Embedded HTML strings throughout

The main `convertANSIToHTML()` function is particularly complex, handling multiple ANSI escape sequence types in a single large function.

## Current State Analysis

### Key Issues

1. **Monolithic ANSI Converter:**
   - `convertANSIToHTML()` spans lines 237-387 (150 lines)
   - Complex state machine for parsing ANSI codes
   - Deep nesting with multiple switch statements
   - Mixed parsing and HTML generation

2. **Repetitive Color Mapping:**
   ```go
   switch code {
   case 30: return "color: #000000" // Black
   case 31: return "color: #ff0000" // Red
   case 32: return "color: #00ff00" // Green
   // ... 20+ more cases
   }
   ```

3. **Mixed Concerns:**
   - ANSI parsing logic mixed with HTML generation
   - Template helper functions mixed with conversion logic
   - No clear separation between parsing and rendering

4. **Embedded HTML Strings:**
   - HTML snippets scattered throughout
   - Difficult to maintain consistent styling
   - No template usage for HTML generation

## Proposed Solution

### Architecture Redesign

```
renderer/
├── html.go           # Main HTML renderer interface
├── ansi/
│   ├── parser.go     # ANSI escape sequence parser
│   ├── converter.go  # ANSI to HTML converter
│   ├── colors.go     # Color mapping and utilities
│   └── state.go      # Parser state management
├── builders/
│   ├── html.go       # HTML element builders
│   ├── attributes.go # HTML attribute helpers
│   └── escape.go     # HTML escaping utilities
└── helpers.go        # Template helper functions
```

### Key Abstractions

1. **ANSIParser Interface:**
   ```go
   type ANSIParser interface {
       Parse(input string) ([]ANSIToken, error)
   }
   
   type ANSIToken struct {
       Type    TokenType // Text, EscapeSequence
       Content string
       Codes   []int     // For escape sequences
   }
   ```

2. **ANSIConverter:**
   ```go
   type ANSIConverter struct {
       colorMap   ColorMapper
       styleMap   StyleMapper
       htmlBuilder HTMLBuilder
   }
   
   func (c *ANSIConverter) Convert(input string) (string, error) {
       tokens, err := c.parser.Parse(input)
       if err != nil {
           return "", err
       }
       return c.renderTokens(tokens)
   }
   ```

3. **HTMLBuilder:**
   ```go
   type HTMLBuilder interface {
       StartSpan(classes []string, styles map[string]string) string
       EndSpan() string
       Text(content string) string
       Build() string
   }
   ```

## Implementation Steps

### Step 1: Extract ANSI Color Mapping (30 min)

Create `internal/renderer/ansi/colors.go`:

```go
package ansi

type ColorMapper struct {
    foregroundColors map[int]string
    backgroundColors map[int]string
    rgb256Palette    []string
}

func NewColorMapper() *ColorMapper {
    return &ColorMapper{
        foregroundColors: initForegroundColors(),
        backgroundColors: initBackgroundColors(),
        rgb256Palette:    init256ColorPalette(),
    }
}

func initForegroundColors() map[int]string {
    return map[int]string{
        30: "#000000", // Black
        31: "#ff0000", // Red
        32: "#00ff00", // Green
        33: "#ffff00", // Yellow
        34: "#0000ff", // Blue
        35: "#ff00ff", // Magenta
        36: "#00ffff", // Cyan
        37: "#ffffff", // White
        // Bright colors
        90: "#808080", // Bright Black
        91: "#ff8080", // Bright Red
        // ... etc
    }
}

func (m *ColorMapper) GetForegroundColor(code int) (string, bool) {
    color, exists := m.foregroundColors[code]
    return color, exists
}

func (m *ColorMapper) Get256Color(index int) string {
    if index >= 0 && index < len(m.rgb256Palette) {
        return m.rgb256Palette[index]
    }
    return "#ffffff" // Default
}
```

### Step 2: Create ANSI Parser (1 hour)

Create `internal/renderer/ansi/parser.go`:

```go
package ansi

import (
    "regexp"
    "strconv"
    "strings"
)

type TokenType int

const (
    TokenText TokenType = iota
    TokenEscapeSequence
)

type ANSIToken struct {
    Type    TokenType
    Content string
    Codes   []int
}

type ANSIParser struct {
    escapeRegex *regexp.Regexp
}

func NewANSIParser() *ANSIParser {
    return &ANSIParser{
        escapeRegex: regexp.MustCompile(`\x1b\[([0-9;]+)m`),
    }
}

func (p *ANSIParser) Parse(input string) ([]ANSIToken, error) {
    var tokens []ANSIToken
    lastEnd := 0
    
    matches := p.escapeRegex.FindAllStringSubmatchIndex(input, -1)
    
    for _, match := range matches {
        // Add text before escape sequence
        if match[0] > lastEnd {
            tokens = append(tokens, ANSIToken{
                Type:    TokenText,
                Content: input[lastEnd:match[0]],
            })
        }
        
        // Parse escape codes
        codesStr := input[match[2]:match[3]]
        codes := p.parseCodes(codesStr)
        
        tokens = append(tokens, ANSIToken{
            Type:  TokenEscapeSequence,
            Codes: codes,
        })
        
        lastEnd = match[1]
    }
    
    // Add remaining text
    if lastEnd < len(input) {
        tokens = append(tokens, ANSIToken{
            Type:    TokenText,
            Content: input[lastEnd:],
        })
    }
    
    return tokens, nil
}

func (p *ANSIParser) parseCodes(codesStr string) []int {
    parts := strings.Split(codesStr, ";")
    codes := make([]int, 0, len(parts))
    
    for _, part := range parts {
        if code, err := strconv.Atoi(part); err == nil {
            codes = append(codes, code)
        }
    }
    
    return codes
}
```

### Step 3: Create ANSI State Manager (45 min)

Create `internal/renderer/ansi/state.go`:

```go
package ansi

type ANSIState struct {
    Bold          bool
    Italic        bool
    Underline     bool
    StrikeThrough bool
    FgColor       string
    BgColor       string
}

func NewANSIState() *ANSIState {
    return &ANSIState{}
}

func (s *ANSIState) ApplyCodes(codes []int, colorMapper *ColorMapper) {
    i := 0
    for i < len(codes) {
        code := codes[i]
        
        switch code {
        case 0: // Reset
            s.Reset()
        case 1: // Bold
            s.Bold = true
        case 3: // Italic
            s.Italic = true
        case 4: // Underline
            s.Underline = true
        case 9: // Strike through
            s.StrikeThrough = true
        case 22: // Normal intensity
            s.Bold = false
        case 23: // Not italic
            s.Italic = false
        case 24: // Not underlined
            s.Underline = false
        case 29: // Not crossed out
            s.StrikeThrough = false
        case 30, 31, 32, 33, 34, 35, 36, 37: // Foreground colors
            if color, ok := colorMapper.GetForegroundColor(code); ok {
                s.FgColor = color
            }
        case 38: // Extended foreground color
            if i+2 < len(codes) && codes[i+1] == 5 {
                s.FgColor = colorMapper.Get256Color(codes[i+2])
                i += 2
            }
        case 40, 41, 42, 43, 44, 45, 46, 47: // Background colors
            if color, ok := colorMapper.GetBackgroundColor(code); ok {
                s.BgColor = color
            }
        case 48: // Extended background color
            if i+2 < len(codes) && codes[i+1] == 5 {
                s.BgColor = colorMapper.Get256Color(codes[i+2])
                i += 2
            }
        }
        i++
    }
}

func (s *ANSIState) Reset() {
    *s = ANSIState{}
}

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
```

### Step 4: Create HTML Builders (45 min)

Create `internal/renderer/builders/html.go`:

```go
package builders

import (
    "html"
    "strings"
)

type HTMLBuilder struct {
    parts []string
}

func NewHTMLBuilder() *HTMLBuilder {
    return &HTMLBuilder{
        parts: make([]string, 0, 100),
    }
}

func (b *HTMLBuilder) StartSpan(classes []string, styles map[string]string) {
    var attrs []string
    
    if len(classes) > 0 {
        attrs = append(attrs, fmt.Sprintf(`class="%s"`, strings.Join(classes, " ")))
    }
    
    if len(styles) > 0 {
        var styleStrs []string
        for key, value := range styles {
            styleStrs = append(styleStrs, fmt.Sprintf("%s: %s", key, value))
        }
        attrs = append(attrs, fmt.Sprintf(`style="%s"`, strings.Join(styleStrs, "; ")))
    }
    
    if len(attrs) > 0 {
        b.parts = append(b.parts, fmt.Sprintf("<span %s>", strings.Join(attrs, " ")))
    } else {
        b.parts = append(b.parts, "<span>")
    }
}

func (b *HTMLBuilder) EndSpan() {
    b.parts = append(b.parts, "</span>")
}

func (b *HTMLBuilder) Text(content string) {
    b.parts = append(b.parts, html.EscapeString(content))
}

func (b *HTMLBuilder) Raw(html string) {
    b.parts = append(b.parts, html)
}

func (b *HTMLBuilder) Build() string {
    return strings.Join(b.parts, "")
}
```

### Step 5: Refactor Main Converter (1 hour)

Create `internal/renderer/ansi/converter.go`:

```go
package ansi

import (
    "github.com/user/cclogviewer/internal/renderer/builders"
)

type ANSIConverter struct {
    parser      *ANSIParser
    colorMapper *ColorMapper
}

func NewANSIConverter() *ANSIConverter {
    return &ANSIConverter{
        parser:      NewANSIParser(),
        colorMapper: NewColorMapper(),
    }
}

func (c *ANSIConverter) ConvertToHTML(input string) (string, error) {
    tokens, err := c.parser.Parse(input)
    if err != nil {
        return "", err
    }
    
    builder := builders.NewHTMLBuilder()
    state := NewANSIState()
    openSpan := false
    
    for _, token := range tokens {
        switch token.Type {
        case TokenEscapeSequence:
            // Close current span if open
            if openSpan {
                builder.EndSpan()
                openSpan = false
            }
            
            // Apply codes to state
            state.ApplyCodes(token.Codes, c.colorMapper)
            
        case TokenText:
            if token.Content == "" {
                continue
            }
            
            // Open span with current state if needed
            classes := state.GetClasses()
            styles := state.GetStyles()
            
            if len(classes) > 0 || len(styles) > 0 {
                builder.StartSpan(classes, styles)
                openSpan = true
            }
            
            builder.Text(token.Content)
        }
    }
    
    // Close final span if open
    if openSpan {
        builder.EndSpan()
    }
    
    return builder.Build(), nil
}
```

### Step 6: Update Main html.go (30 min)

Refactor the main html.go file to use new components:

```go
package renderer

import (
    "github.com/user/cclogviewer/internal/renderer/ansi"
    "github.com/user/cclogviewer/internal/renderer/builders"
)

var ansiConverter = ansi.NewANSIConverter()

// ConvertANSIToHTML converts ANSI escape sequences to HTML
func ConvertANSIToHTML(input string) string {
    html, err := ansiConverter.ConvertToHTML(input)
    if err != nil {
        // Fallback to escaped text
        return builders.EscapeHTML(input)
    }
    return html
}

// Template functions remain but use new utilities
func templateFuncs() template.FuncMap {
    return template.FuncMap{
        "convertANSI":    ConvertANSIToHTML,
        "escapeHTML":     builders.EscapeHTML,
        "formatPath":     formatPath,
        "formatDuration": formatDuration,
        // ... other template functions
    }
}
```

## Benefits

1. **Separation of Concerns:** ANSI parsing separate from HTML generation
2. **Testability:** Each component can be tested independently
3. **Maintainability:** Smaller, focused functions
4. **Extensibility:** Easy to add new ANSI codes or output formats
5. **Performance:** More efficient parsing with dedicated parser

## Risks and Mitigation

1. **Risk:** Breaking ANSI rendering for edge cases
   - **Mitigation:** Comprehensive test suite with real-world examples
   - **Mitigation:** Visual regression testing

2. **Risk:** Performance regression from abstraction
   - **Mitigation:** Benchmark parsing performance
   - **Mitigation:** Use string builder for efficiency

3. **Risk:** Incorrect color mapping
   - **Mitigation:** Validate against standard ANSI color tables
   - **Mitigation:** Test with various terminal outputs

## Success Criteria

- [ ] `convertANSIToHTML()` reduced to <20 lines
- [ ] No function exceeds 50 lines
- [ ] ANSI parsing completely separate from HTML generation
- [ ] 95%+ test coverage for ANSI components
- [ ] No visual changes to rendered output
- [ ] Performance within 10% of original

## Future Enhancements

This refactoring enables:
- Support for additional ANSI features (256/true color, cursor control)
- Alternative output formats (Markdown, plain text)
- Configurable color themes
- Terminal emulation features
- ANSI art rendering support