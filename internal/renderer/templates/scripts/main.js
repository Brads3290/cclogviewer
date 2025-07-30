// Main JavaScript functionality for Claude Code Log Viewer

// Debug logging function (will be defined by Go template)
// const debugLog = (...args) => console.log('[DEBUG]', ...args);

// Use event delegation for tool call toggling
document.addEventListener('click', (e) => {
    // Handle tool header clicks
    const toolHeader = e.target.closest('.tool-header');
    if (toolHeader) {
        e.preventDefault();
        e.stopPropagation();
        
        if (typeof debugLog !== 'undefined') {
            debugLog('=== Tool header clicked ===');
            const toolCall = toolHeader.parentElement;
            debugLog('Tool debug-id:', toolCall.getAttribute('data-debug-id'));
            debugLog('Tool name:', toolCall.getAttribute('data-tool-name'));
            debugLog('Parent entry:', toolCall.getAttribute('data-parent-entry'));
            debugLog('Has task entries:', toolCall.getAttribute('data-has-task-entries'));
            debugLog('Current classes:', toolCall.className);
            debugLog('Has expanded class:', toolCall.classList.contains('expanded'));
            
            // Build hierarchy path
            const buildPath = (elem) => {
                const path = [];
                let current = elem;
                while (current) {
                    const debugId = current.getAttribute('data-debug-id');
                    if (debugId) {
                        path.unshift(debugId);
                    }
                    current = current.parentElement.closest('[data-debug-id]');
                }
                return path.join(' > ');
            };
            debugLog('Full hierarchy path:', buildPath(toolCall));
            
            // Check if it's in a task-entry
            const taskEntry = toolCall.closest('.task-entry');
            debugLog('Inside task-entry:', !!taskEntry);
            if (taskEntry) {
                debugLog('Task entry debug-id:', taskEntry.getAttribute('data-debug-id'));
                debugLog('Parent tool:', taskEntry.getAttribute('data-parent-tool'));
                debugLog('Nested task entries within:', 
                    Array.from(toolCall.querySelectorAll('.task-entry')).length);
            }
            
            // Get tool-details element
            const toolDetails = toolCall.querySelector('.tool-details');
            if (toolDetails) {
                debugLog('Tool details element:', toolDetails);
                debugLog('Tool details computed style:', 
                    window.getComputedStyle(toolDetails).display);
                
                // Check all applicable CSS rules
                const allRules = [];
                for (const sheet of document.styleSheets) {
                    try {
                        for (const rule of sheet.cssRules) {
                            if (rule.selectorText && toolDetails.matches(rule.selectorText)) {
                                allRules.push({
                                    selector: rule.selectorText,
                                    display: rule.style.display
                                });
                            }
                        }
                    } catch (e) {
                        // Cross-origin stylesheets will throw
                    }
                }
                debugLog('Matching CSS rules:', allRules);
            }
        }
        
        toolHeader.parentElement.classList.toggle('expanded');
        
        if (typeof debugLog !== 'undefined') {
            debugLog('=== After toggle ===');
            debugLog('Has expanded class:', 
                toolHeader.parentElement.classList.contains('expanded'));
            if (toolDetails) {
                const afterDisplay = window.getComputedStyle(toolDetails).display;
                debugLog('Tool details computed style:', afterDisplay);
                debugLog('Display changed:', afterDisplay !== 'none' ? 'VISIBLE' : 'HIDDEN');
            }
            debugLog('========================');
        }
    }
    
    // Handle result header clicks
    const resultHeader = e.target.closest('.result-header');
    if (resultHeader) {
        e.preventDefault();
        e.stopPropagation();
        const icon = resultHeader.querySelector('.result-expand-icon');
        const content = resultHeader.nextElementSibling;
        if (content) {
            const isHidden = content.style.display === 'none';
            content.style.display = isHidden ? 'block' : 'none';
            icon.style.transform = isHidden ? 'rotate(90deg)' : 'rotate(0deg)';
        }
    }
    
    // Handle caveat message header clicks
    const caveatHeader = e.target.closest('.caveat-header');
    if (caveatHeader) {
        e.preventDefault();
        e.stopPropagation();
        const icon = caveatHeader.querySelector('.caveat-expand-icon');
        const content = caveatHeader.nextElementSibling;
        if (content) {
            const isHidden = content.style.display === 'none';
            content.style.display = isHidden ? 'block' : 'none';
            icon.style.transform = isHidden ? 'rotate(90deg)' : 'rotate(0deg)';
            
            // No need to update the header text - it stays the same
        }
    }
    
});

// Global state for token details visibility
let tokenDetailsExpanded = false;

// Handle token details toggle
document.addEventListener('click', (e) => {
    const tokenToggle = e.target.closest('.token-toggle');
    if (tokenToggle) {
        e.preventDefault();
        e.stopPropagation();
        
        // Toggle global state
        tokenDetailsExpanded = !tokenDetailsExpanded;
        
        // Update all token toggles
        const allTokenToggles = document.querySelectorAll('.token-toggle');
        allTokenToggles.forEach(toggle => {
            const details = toggle.querySelector('.token-details');
            const icon = toggle.querySelector('.token-expand-icon');
            
            if (tokenDetailsExpanded) {
                details.classList.add('show');
                icon.textContent = '[-]';
            } else {
                details.classList.remove('show');
                icon.textContent = '[+]';
            }
        });
    }
});

// Debug CSS rules on page load
if (typeof debugLog !== 'undefined') {
    document.addEventListener('DOMContentLoaded', () => {
        debugLog('=== Page loaded - checking CSS rules ===');
        const toolCalls = document.querySelectorAll('.tool-call');
        debugLog('Total tool calls found:', toolCalls.length);
        
        // List all tool calls with their debug IDs
        toolCalls.forEach((tc, index) => {
            const debugId = tc.getAttribute('data-debug-id');
            const toolName = tc.getAttribute('data-tool-name');
            const hasTaskEntries = tc.getAttribute('data-has-task-entries');
            debugLog('Tool ' + index + ': ' + debugId + ' (' + toolName + ') - Has tasks: ' + hasTaskEntries);
        });
        
        // Check nested tool calls
        const nestedToolCalls = document.querySelectorAll('.task-entry .tool-call');
        debugLog('\nNested tool calls:', nestedToolCalls.length);
        
        // Check for any expanded ancestors
        debugLog('\nChecking for expanded ancestors:');
        document.querySelectorAll('.expanded').forEach((elem, index) => {
            debugLog('Expanded element ' + index + ':', elem.getAttribute('data-debug-id') || elem.className);
        });
        
        // Detail each nested tool call
        nestedToolCalls.forEach((tc, index) => {
            const debugId = tc.getAttribute('data-debug-id');
            const toolName = tc.getAttribute('data-tool-name');
            const taskEntry = tc.closest('.task-entry');
            const taskDebugId = taskEntry ? taskEntry.getAttribute('data-debug-id') : 'none';
            const toolDetails = tc.querySelector('.tool-details');
            const display = toolDetails ? window.getComputedStyle(toolDetails).display : 'no-details';
            
            debugLog('Nested ' + index + ': ' + debugId + ' (' + toolName + ')');
            debugLog('  In task-entry: ' + taskDebugId);
            debugLog('  Tool-details display: ' + display);
            debugLog('  Has expanded class: ' + tc.classList.contains('expanded'));
        });
        
        debugLog('========================');
    });
}