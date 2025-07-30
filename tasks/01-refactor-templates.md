# Task 01: Refactor templates.go

**File:** `internal/renderer/templates.go`  
**Current Size:** 1,110 lines  
**Priority:** HIGH  
**Estimated Effort:** 4-6 hours  

## Problem Summary

The entire HTML template, CSS styling, and JavaScript functionality is embedded as a single 1,100+ line string literal in the Go source code. This makes the template:
- Difficult to maintain and edit
- Impossible to preview without compiling
- Hard to debug JavaScript issues
- Challenging to apply CSS changes
- Prone to syntax errors that only appear at runtime

## Current State Analysis

The file contains:
- **Lines 10-1120:** A single `htmlTemplate` const string containing:
  - Full HTML document structure
  - ~400 lines of embedded CSS styles
  - ~200 lines of embedded JavaScript
  - Go template directives mixed throughout

Key issues:
1. No syntax highlighting for HTML/CSS/JS in most editors
2. No ability to use web development tools
3. Changes require recompiling the entire binary
4. No code reuse or modularity
5. Difficult to unit test individual components

## Proposed Solution

### Phase 1: Extract and Embed Files

1. **Create template file structure:**
   ```
   internal/renderer/templates/
   ├── base.html
   ├── styles/
   │   ├── main.css
   │   ├── components.css
   │   └── themes.css
   ├── scripts/
   │   ├── main.js
   │   ├── search.js
   │   └── navigation.js
   └── partials/
       ├── entry.html
       ├── tool-call.html
       └── sidebar.html
   ```

2. **Use Go 1.16+ embed directive:**
   ```go
   //go:embed templates/*
   var templateFS embed.FS
   ```

3. **Split the monolithic template:**
   - Extract CSS into separate stylesheets
   - Extract JavaScript into modules
   - Create reusable HTML partials for entries, tools, etc.

### Phase 2: Implement Template Loading

1. **Create template loader function:**
   ```go
   func loadTemplates() (*template.Template, error) {
       tmpl := template.New("base")
       
       // Load all template files
       err := loadTemplateFiles(tmpl, templateFS, "templates")
       if err != nil {
           return nil, err
       }
       
       return tmpl, nil
   }
   ```

2. **Add template caching for production**
3. **Support development mode with hot reload**

### Phase 3: Refactor Component Structure

1. **Extract reusable components:**
   - Entry display component
   - Tool call formatter component
   - Search functionality component
   - Navigation sidebar component

2. **Create template functions for common operations:**
   - Time formatting
   - Token display
   - Status indicators
   - Role badges

## Implementation Steps

### Step 1: Create Directory Structure (30 min)
```bash
mkdir -p internal/renderer/templates/{styles,scripts,partials}
```

### Step 2: Extract HTML Structure (1 hour)
1. Copy current template to `templates/base.html`
2. Identify and extract repeated HTML patterns to partials:
   - Entry rendering logic → `partials/entry.html`
   - Tool call display → `partials/tool-call.html`
   - Sidebar navigation → `partials/sidebar.html`

### Step 3: Extract CSS (1 hour)
1. Move all CSS from `<style>` tag to separate files:
   - Core styles → `styles/main.css`
   - Component-specific styles → `styles/components.css`
   - Theme variables → `styles/themes.css`
2. Add CSS minification in build process (optional)

### Step 4: Extract JavaScript (1.5 hours)
1. Move all JavaScript to separate files:
   - Core functionality → `scripts/main.js`
   - Search implementation → `scripts/search.js`
   - Navigation logic → `scripts/navigation.js`
2. Consider using ES6 modules if browser support allows
3. Add JavaScript minification in build process (optional)

### Step 5: Update Go Code (1.5 hours)
1. Add embed directives
2. Implement template loader
3. Update `Render()` function to use new template structure
4. Add error handling for missing templates
5. Update tests

### Step 6: Testing and Validation (30 min)
1. Verify all functionality works as before
2. Test with various log files
3. Check browser compatibility
4. Validate HTML/CSS/JS syntax

## Benefits

1. **Maintainability:** Easy to edit templates with proper syntax highlighting
2. **Development Experience:** Can use web development tools and linters
3. **Performance:** Templates cached at compile time, no runtime parsing
4. **Modularity:** Reusable components and clear separation of concerns
5. **Testing:** Can unit test individual components
6. **Debugging:** Browser dev tools work properly with separate files

## Risks and Mitigation

1. **Risk:** Breaking existing functionality
   - **Mitigation:** Comprehensive testing, keep old template as fallback

2. **Risk:** Increased complexity in build process
   - **Mitigation:** Use standard Go embed, no external tools needed

3. **Risk:** Performance regression
   - **Mitigation:** Templates are embedded at compile time, same performance

## Success Criteria

- [ ] All HTML/CSS/JS extracted to separate files
- [ ] Templates load correctly using embed.FS
- [ ] No functional regressions
- [ ] Improved developer experience when editing templates
- [ ] Clear component structure for future enhancements
- [ ] All tests pass

## Future Enhancements

Once this refactoring is complete, it enables:
- Easy theme switching
- Component library for consistent UI
- Template customization without recompiling
- A/B testing different layouts
- Internationalization support