package renderer

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Claude Code Log Viewer</title>
    <style>
        * {
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            margin: 0;
            padding: 20px;
        }
        
        .container {
            max-width: 1200px;
            margin: 0 auto;
            background: white;
            border-radius: 8px;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            padding: 20px;
        }
        
        h1 {
            color: #2c3e50;
            border-bottom: 2px solid #3498db;
            padding-bottom: 10px;
            margin-bottom: 20px;
        }
        
        .entry {
            margin-bottom: 20px;
            border-left: 3px solid transparent;
            padding-left: 15px;
            transition: all 0.2s ease;
        }
        
        .entry {
            padding: 10px 15px;
            border-radius: 4px;
        }
        
        /* Color schemes for different depth levels */
        /* Depth 1 - Root conversation: Blue for user, Green for assistant */
        .entry.depth-1.user {
            background: #e3f2fd;
            border-left-color: #1976d2;
        }
        
        .entry.depth-1.assistant {
            background: #e8f5e9;
            border-left-color: #388e3c;
        }
        
        /* Depth 2 - Purple theme */
        .entry.depth-2.user {
            background: #f3e5f5;
            border-left-color: #7b1fa2;
        }
        
        .entry.depth-2.assistant {
            background: #ede7f6;
            border-left-color: #512da8;
        }
        
        /* Depth 3 - Orange theme */
        .entry.depth-3.user {
            background: #fff3e0;
            border-left-color: #f57c00;
        }
        
        .entry.depth-3.assistant {
            background: #fbe9e7;
            border-left-color: #d84315;
        }
        
        /* Depth 4 - Teal theme */
        .entry.depth-4.user {
            background: #e0f2f1;
            border-left-color: #00796b;
        }
        
        .entry.depth-4.assistant {
            background: #e0f7fa;
            border-left-color: #00838f;
        }
        
        /* Depth 5 - Pink theme */
        .entry.depth-5.user {
            background: #fce4ec;
            border-left-color: #c2185b;
        }
        
        .entry.depth-5.assistant {
            background: #f8bbd0;
            border-left-color: #ad1457;
        }
        
        /* Special styling: sidechain user messages use parent assistant colors */
        .entry.sidechain.depth-2.user {
            /* Use depth-1 assistant colors */
            background: #e8f5e9;
            border-left-color: #388e3c;
        }
        
        .entry.sidechain.depth-3.user {
            /* Use depth-2 assistant colors */
            background: #ede7f6;
            border-left-color: #512da8;
        }
        
        .entry.sidechain.depth-4.user {
            /* Use depth-3 assistant colors */
            background: #fbe9e7;
            border-left-color: #d84315;
        }
        
        .entry.sidechain.depth-5.user {
            /* Use depth-4 assistant colors */
            background: #e0f7fa;
            border-left-color: #00838f;
        }
        
        .entry.sidechain.depth-1.user {
            /* Use depth-5 assistant colors (wraps around) */
            background: #f8bbd0;
            border-left-color: #ad1457;
        }
        
        .task-entry .entry.sidechain {
            margin-left: 0; /* Override any margin for entries within task */
        }
        
        
        .entry-header {
            display: flex;
            align-items: center;
            gap: 10px;
            margin-bottom: 8px;
            font-size: 0.9em;
            color: #666;
        }
        
        .role {
            font-weight: bold;
            text-transform: capitalize;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 0.85em;
            color: white;
        }
        
        /* Role labels inherit colors from their depth */
        .entry.depth-1 .role.user {
            background: #1976d2;
        }
        
        .entry.depth-1 .role.assistant {
            background: #388e3c;
        }
        
        .entry.depth-2 .role.user {
            background: #7b1fa2;
        }
        
        .entry.depth-2 .role.assistant,
        .entry.depth-2 .role.subagent {
            background: #512da8;
        }
        
        .entry.depth-3 .role.user {
            background: #f57c00;
        }
        
        .entry.depth-3 .role.assistant,
        .entry.depth-3 .role.subagent {
            background: #d84315;
        }
        
        .entry.depth-4 .role.user {
            background: #00796b;
        }
        
        .entry.depth-4 .role.assistant,
        .entry.depth-4 .role.subagent {
            background: #00838f;
        }
        
        .entry.depth-5 .role.user {
            background: #c2185b;
        }
        
        .entry.depth-5 .role.assistant,
        .entry.depth-5 .role.subagent {
            background: #ad1457;
        }
        
        /* Special case: sidechain user messages use parent assistant label colors */
        .entry.sidechain.depth-2.user .role.assistant {
            background: #388e3c; /* depth-1 assistant color */
        }
        
        .entry.sidechain.depth-3.user .role.assistant {
            background: #512da8; /* depth-2 assistant color */
        }
        
        .entry.sidechain.depth-4.user .role.assistant {
            background: #d84315; /* depth-3 assistant color */
        }
        
        .entry.sidechain.depth-5.user .role.assistant {
            background: #00838f; /* depth-4 assistant color */
        }
        
        .timestamp {
            color: #999;
            font-size: 0.85em;
        }
        
        .content {
            white-space: pre-wrap;
            word-wrap: break-word;
        }
        
        .tool-calls {
            margin-top: 10px;
        }
        
        .tool-call {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 10px;
            margin-bottom: 10px;
        }
        
        .tool-header {
            display: flex;
            align-items: center;
            gap: 10px;
            cursor: pointer;
            user-select: none;
            position: relative;
            padding: 10px;
            margin: -10px;
            border-radius: 4px;
            padding-right: 150px; /* Make room for the tool ID */
        }
        
        .tool-header:hover {
            background: #e9ecef;
        }
        
        .tool-name {
            font-weight: bold;
            color: #495057;
        }
        
        .tool-description {
            color: #6c757d;
            font-size: 0.9em;
        }
        
        .tool-id {
            position: absolute;
            right: 10px;
            color: #999;
            font-size: 0.85em;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        .tool-id-copy {
            margin-bottom: 10px;
            color: #666;
            font-size: 0.9em;
        }
        
        .expand-icon {
            width: 20px;
            height: 20px;
            transition: transform 0.2s;
        }
        
        .tool-call.expanded > .tool-header > .expand-icon {
            transform: rotate(90deg);
        }
        
        .tool-details {
            display: none;
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px solid #dee2e6;
        }
        
        /* Only show tool-details when the immediate parent tool-call has expanded class */
        .tool-call.expanded > .tool-details {
            display: block;
        }
        
        .tool-input {
            background: #f1f3f5;
            padding: 10px;
            border-radius: 4px;
            overflow-wrap: break-word;
            word-wrap: break-word;
            word-break: break-word;
            font-size: 1em;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            line-height: 1.4;
        }
        
        .tool-input pre {
            white-space: pre-wrap;
            word-wrap: break-word;
            word-break: break-word;
            overflow-wrap: break-word;
        }
        
        .tool-result {
            background: #e3f2fd;
            padding: 10px;
            border-radius: 4px;
            margin-top: 5px;
            white-space: pre-wrap;
        }
        
        .tool-result.error {
            background: #ffebee;
            color: #c62828;
            border: 1px solid #ef5350;
        }
        
        .children {
            margin-left: 20px;
            margin-top: 15px;
            padding-left: 20px;
            border-left: 2px dashed #ddd;
        }
        
        .task-children {
            background: #fafafa;
            border-radius: 4px;
            padding: 10px;
            margin-top: 10px;
        }
        
        pre {
            margin: 0;
            padding: 0;
            background: transparent;
            border: none;
        }
        
        code {
            background: #f4f4f4;
            padding: 2px 4px;
            border-radius: 3px;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        pre code {
            background: none;
            padding: 0;
        }
        
        /* Diff view styles */
        .diff-content {
            background: #fafafa;
        }
        
        .diff-content.unified {
            padding: 0;
        }
        
        .diff-code {
            margin: 0;
            padding: 0;
            overflow-x: auto;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.85em;
            white-space: pre-wrap;
            word-wrap: break-word;
            background: #fafafa;
        }
        
        .diff-line {
            display: flex;
            align-items: flex-start;
            padding: 0;
            margin: 0;
        }
        
        .diff-line.line-removed {
            background: #ffebee;
        }
        
        .diff-line.line-added {
            background: #e8f5e9;
        }
        
        .diff-line.line-unchanged {
            background: transparent;
        }
        
        .diff-code .line-number {
            color: #999;
            user-select: none;
            margin-right: 10px;
            display: inline-block;
            text-align: right;
            flex-shrink: 0;
            padding-left: 10px;
            line-height: 1.3;
        }
        
        .diff-code .line-prefix {
            font-weight: bold;
            margin-right: 8px;
            flex-shrink: 0;
            line-height: 1.3;
        }
        
        .diff-line.line-removed .line-prefix,
        .diff-line.line-removed .line-content {
            color: #d32f2f;
        }
        
        .diff-line.line-added .line-prefix,
        .diff-line.line-added .line-content {
            color: #388e3c;
        }
        
        .diff-line.line-unchanged .line-prefix,
        .diff-line.line-unchanged .line-content {
            color: #666;
        }
        
        .diff-code .line-content {
            flex: 1;
            white-space: pre-wrap;
            word-wrap: break-word;
            padding-right: 10px;
            line-height: 1.3;
        }
        
        
        /* Compact todo list styles */
        .todo-compact {
            margin-top: 10px;
            background: #f8f9fa;
            border: 1px solid #e9ecef;
            border-radius: 4px;
            padding: 8px 12px;
            font-size: 0.85em;
        }
        
        .todo-compact-summary {
            display: flex;
            align-items: center;
            gap: 12px;
            margin-bottom: 6px;
        }
        
        .todo-compact-title {
            font-weight: 600;
            color: #495057;
        }
        
        .todo-stat {
            display: inline-flex;
            align-items: center;
            gap: 3px;
            padding: 2px 6px;
            border-radius: 3px;
            font-size: 0.8em;
        }
        
        .todo-stat.completed {
            background: #d4edda;
            color: #155724;
        }
        
        .todo-stat.in-progress {
            background: #fff3cd;
            color: #856404;
        }
        
        .todo-stat.pending {
            background: #e2e3e5;
            color: #383d41;
        }
        
        .todo-compact-items {
            margin-left: 20px;
        }
        
        .todo-compact-item {
            padding: 3px 0;
            color: #495057;
            display: flex;
            align-items: center;
            gap: 6px;
        }
        
        .todo-compact-item.completed {
            opacity: 0.6;
            text-decoration: line-through;
        }
        
        .todo-icon {
            width: 16px;
            text-align: center;
            flex-shrink: 0;
        }
        
        .todo-priority-badge {
            display: inline-block;
            width: 14px;
            height: 14px;
            line-height: 14px;
            text-align: center;
            border-radius: 2px;
            font-size: 0.7em;
            font-weight: bold;
            margin-left: 4px;
        }
        
        .todo-priority-badge.high {
            background: #dc3545;
            color: white;
        }
        
        .todo-priority-badge.medium {
            background: #ffc107;
            color: #333;
        }
        
        /* Token details toggle styles */
        .token-toggle {
            cursor: pointer;
            user-select: none;
            font-size: 0.85em;
            color: #666;
        }
        
        .token-toggle:hover {
            color: #333;
        }
        
        .token-details {
            display: none;
            color: #999;
            font-size: 0.85em;
            margin-left: 5px;
        }
        
        .token-details.show {
            display: inline;
        }
        
        .token-expand-icon {
            display: inline-block;
            font-family: monospace;
        }
        
        /* Read tool display styles */
        .read-display {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 0;
            margin: 0;
            overflow: hidden;
        }
        
        .read-header {
            background: #e9ecef;
            padding: 10px 15px;
            font-weight: bold;
            border-bottom: 1px solid #dee2e6;
        }
        
        .read-header .file-path {
            color: #0056b3;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9em;
        }
        
        .read-header .line-info {
            color: #6c757d;
            font-size: 0.8em;
            margin-left: 10px;
            font-weight: normal;
        }
        
        .read-content {
            background: #fafafa;
            padding: 0;
        }
        
        .read-code {
            margin: 0;
            padding: 0;
            overflow-x: auto;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.85em;
            line-height: 1.4;
            white-space: pre-wrap;
            word-wrap: break-word;
            background: #fafafa;
            /* Ensure proper UTF-8 rendering */
            unicode-bidi: plaintext;
        }
        
        .read-line {
            display: flex;
            padding: 0;
            margin: 0;
            line-height: 1.4;
        }
        
        .read-line:hover {
            background: #f0f0f0;
        }
        
        .read-code .line-number {
            color: #999;
            user-select: none;
            margin-right: 15px;
            display: inline-block;
            text-align: right;
            width: 40px;
            flex-shrink: 0;
            align-self: flex-start;
            padding-left: 10px;
        }
        
        .read-code .line-content {
            flex: 1;
            white-space: pre-wrap;
            word-wrap: break-word;
            padding-right: 10px;
        }
        
        /* Bash tool display styles */
        .bash-display {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 6px;
            padding: 0;
            margin: 0;
            overflow: hidden;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        }
        
        .bash-header {
            background: #e9ecef;
            padding: 8px 12px;
            border-bottom: 1px solid #dee2e6;
            display: flex;
            align-items: center;
            gap: 12px;
        }
        
        .bash-header .terminal-icon {
            color: #0d6efd;
            font-size: 1.1em;
        }
        
        .bash-header .command-label {
            color: #495057;
            font-size: 0.9em;
            font-weight: 600;
        }
        
        .bash-header .description {
            color: #6c757d;
            font-size: 0.85em;
            margin-left: auto;
        }
        
        .bash-terminal {
            background: #f8f9fa;
            padding: 12px 16px;
            font-size: 0.9em;
            line-height: 1.4;
            color: #212529;
        }
        
        .bash-cwd {
            color: #6c757d;
            font-size: 0.85em;
            margin-bottom: 4px;
        }
        
        .bash-command-line {
            display: flex;
            align-items: flex-start;
            margin-bottom: 8px;
        }
        
        .bash-prompt {
            color: #0d6efd;
            font-weight: bold;
            margin-right: 8px;
            flex-shrink: 0;
        }
        
        .bash-command {
            white-space: pre-wrap;
            word-wrap: break-word;
            color: #212529;
            flex: 1;
        }
        
        .bash-output {
            color: #495057;
            white-space: pre-wrap;
            word-wrap: break-word;
            margin-top: 8px;
            padding-top: 8px;
            border-top: 1px solid #e9ecef;
        }
        
        .bash-timeout {
            color: #fd7e14;
            font-size: 0.8em;
            float: right;
            opacity: 0.7;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>Claude Code Conversation Log</h1>
        {{range .Entries}}
            {{template "entry" .}}
        {{end}}
    </div>
    
    <script>
        {{if $.Debug}}
        const debugLog = (...args) => console.log('[DEBUG]', ...args);
        {{else}}
        const debugLog = () => {};
        {{end}}
        
        // Use event delegation for tool call toggling
        document.addEventListener('click', (e) => {
            // Handle tool header clicks
            const toolHeader = e.target.closest('.tool-header');
            if (toolHeader) {
                e.preventDefault();
                e.stopPropagation();
                
                {{if $.Debug}}
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
                {{end}}
                
                toolHeader.parentElement.classList.toggle('expanded');
                
                {{if $.Debug}}
                debugLog('=== After toggle ===');
                debugLog('Has expanded class:', 
                    toolHeader.parentElement.classList.contains('expanded'));
                if (toolDetails) {
                    const afterDisplay = window.getComputedStyle(toolDetails).display;
                    debugLog('Tool details computed style:', afterDisplay);
                    debugLog('Display changed:', afterDisplay !== 'none' ? 'VISIBLE' : 'HIDDEN');
                }
                debugLog('========================');
                {{end}}
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
        
        {{if $.Debug}}
        // Debug CSS rules on page load
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
        {{end}}
        
    </script>
</body>
</html>

{{define "entry"}}
{{if or (ne .Content "") .ToolCalls}}{{/* Render if content is not empty OR has tool calls */}}
<div class="entry {{.Type}} depth-{{mod (sub .Depth 1) 5 | add 1}}{{if .IsSidechain}} sidechain{{end}}" 
     data-debug-id="entry-{{shortUUID .UUID}}"
     data-uuid="{{.UUID}}"
     data-parent-uuid="{{.ParentUUID}}"
     data-is-sidechain="{{.IsSidechain}}"
     data-depth="{{.Depth}}"
     data-color-depth="{{mod (sub .Depth 1) 5 | add 1}}">
    <div class="entry-header">
        {{if .IsSidechain}}
            {{if eq .Role "user"}}
            <span class="role assistant">Prompt</span>
            {{else if eq .Role "assistant"}}
            <span class="role subagent">Sub Agent</span>
            {{end}}
        {{else}}
            <span class="role {{.Role}}">{{.Role}}</span>
        {{end}}
        <span class="timestamp">{{.Timestamp}}</span>
        {{if .IsSidechain}}
        <span style="color: #9c27b0; font-size: 0.85em;">📎 Task</span>
        {{end}}
        {{if eq .Role "assistant"}}
        <span class="token-toggle">
            {{if .TotalTokens}}Conversation size: {{formatNumber .TotalTokens}} | {{end}}
            <span class="token-expand-icon">[+]</span>
            <span class="token-details">
                {{if or .InputTokens .OutputTokens .CacheReadTokens .CacheCreationTokens}}
                    {{if .InputTokens}}{{formatNumber .InputTokens}} input{{end}}
                    {{if and .InputTokens (or .OutputTokens .CacheReadTokens .CacheCreationTokens)}} | {{end}}
                    {{if .OutputTokens}}~{{formatNumber .OutputTokens}} output{{end}}
                    {{if and .OutputTokens (or .CacheReadTokens .CacheCreationTokens)}} | {{end}}
                    {{if .CacheReadTokens}}{{formatNumber .CacheReadTokens}} cache read{{end}}
                    {{if and .CacheReadTokens .CacheCreationTokens}} | {{end}}
                    {{if .CacheCreationTokens}}{{formatNumber .CacheCreationTokens}} cache write{{end}}
                {{else if .TokenCount}}
                    ~{{formatNumber .TokenCount}} tokens
                {{end}}
            </span>
        </span>
        {{else if eq .Role "user"}}
        <span style="color: #666; font-size: 0.85em;">
            {{if .OutputTokens}}~{{formatNumber .OutputTokens}} tokens{{end}}
        </span>
        {{end}}
    </div>
    
    {{if .IsCaveatMessage}}
    <div class="caveat-message">
        <div class="caveat-header" style="cursor: pointer; user-select: none; display: flex; align-items: center; gap: 5px; color: #999; font-style: italic;">
            <svg class="caveat-expand-icon" width="16" height="16" viewBox="0 0 20 20" fill="currentColor" style="transition: transform 0.2s;">
                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd"></path>
            </svg>
            <span>Command caveat message</span>
        </div>
        <div class="caveat-content" style="display: none; margin-top: 10px;">
            <div class="content">{{formatContent .Content}}</div>
        </div>
    </div>
    {{else if .IsCommandMessage}}
    <div class="command-message">
        <div style="color: #999; font-style: italic;">
            {{.CommandName}}{{if .CommandArgs}} {{.CommandArgs}}{{end}}
        </div>
        {{if .CommandOutput}}
        <div style="margin-top: 5px; color: #666;">
            {{formatContent .CommandOutput}}
        </div>
        {{end}}
    </div>
    {{else if eq .Content ""}}
    {{/* Hide entries with empty content (stdout messages that were linked to commands) */}}
    {{else}}
    <div class="content">{{formatContent .Content}}</div>
    {{end}}
    
    {{if .ToolCalls}}
    <div class="tool-calls">
        {{range .ToolCalls}}
        <div class="tool-call" 
             data-debug-id="tool-{{.ID}}" 
             data-tool-name="{{.Name}}"
             data-parent-entry="{{shortUUID $.UUID}}"
             {{if .TaskEntries}}data-has-task-entries="true"{{end}}>
            {{if eq .Name "Bash"}}
            {{/* For Bash tool, show terminal directly without collapsible section */}}
            <div class="bash-tool-container">
                {{formatBashResult .}}
                <div class="tool-id-copy" style="margin-top: 10px;">Tool ID: <code>{{.ID}}</code></div>
            </div>
            {{else}}
            {{/* For other tools, keep the collapsible section */}}
            <div class="tool-header">
                <svg class="expand-icon" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="tool-name">{{.Name}}</span>
                {{if .Description}}
                <span class="tool-description">{{.Description}}</span>
                {{end}}
                {{if .IsInterrupted}}
                <span style="color: #dc3545; margin-left: 10px;" title="Request interrupted by user">⚠️ Interrupted</span>
                {{end}}
                {{if or .HasMissingResult .HasMissingSidechain}}
                <span style="color: #ffc107; margin-left: 10px;" title="The log file may be incomplete">
                    ⚠️ {{if and .HasMissingResult .HasMissingSidechain}}The tool result and conversation are missing{{else if .HasMissingResult}}The tool result is missing{{else}}The conversation is missing{{end}}. The log file may be incomplete.
                </span>
                {{end}}
                <span class="tool-id">{{.ID}}</span>
            </div>
            <div class="tool-details">
                {{.Input}}
                {{if .TaskEntries}}
                {{$toolID := .ID}}
                <div style="margin-top: 15px;">
                    {{range .TaskEntries}}
                        <div class="task-entry" data-debug-id="task-entry-{{shortUUID .UUID}}"  data-parent-tool="{{$toolID}}">
                            {{template "entry" .}}
                        </div>
                    {{end}}
                </div>
                {{end}}
                {{if .Result}}
                {{if eq .Name "Read"}}
                    {{/* For Read tool, show the content inline */}}
                    {{formatReadResult .Result.Content}}
                {{else if eq .Name "Bash"}}
                    {{/* For Bash tool, result is already integrated into the display */}}
                {{else if and (or (eq .Name "Edit") (eq .Name "MultiEdit")) (not .Result.IsError)}}
                    {{/* For Edit/MultiEdit tools, only show result if it's an error */}}
                {{else}}
                    {{/* For other tools or Edit/MultiEdit with errors, show collapsible result section */}}
                    <div class="tool-result-section" style="margin-top: 15px;">
                        <div class="result-header" style="cursor: pointer; user-select: none; display: flex; align-items: center; gap: 5px;">
                            <svg class="result-expand-icon" width="16" height="16" viewBox="0 0 20 20" fill="currentColor" style="transition: transform 0.2s;">
                                <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                            </svg>
                            <strong>Result</strong>
                        </div>
                        <div class="result-content" style="display: none; margin-top: 10px;">
                            {{formatContent .Result.Content}}
                        </div>
                    </div>
                {{end}}
                {{end}}
                <div class="tool-id-copy" style="margin-top: 10px;">Tool ID: <code>{{.ID}}</code></div>
            </div>
            {{end}}
        </div>
        {{if .CompactView}}
        {{.CompactView}}
        {{end}}
        {{end}}
    </div>
    {{end}}
    
</div>
{{end}}{{/* End of content check */}}
{{end}}`
