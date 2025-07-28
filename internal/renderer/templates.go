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
        
        .entry.user {
            border-left-color: #3498db;
            background: #f0f7ff;
            padding: 10px 15px;
            border-radius: 4px;
        }
        
        .entry.assistant {
            border-left-color: #27ae60;
            background: #f0fff4;
            padding: 10px 15px;
            border-radius: 4px;
        }
        
        .entry.sidechain {
            opacity: 0.9;
        }
        
        .entry.sidechain.user {
            background: #f0fff4; /* Same green background as main assistant */
            border-left-color: #27ae60; /* Green border like main assistant */
        }
        
        .entry.sidechain.assistant {
            background: #fff3cd; /* Yellow background for sub-agent */
            border-left-color: #f39c12; /* Orange border for sub-agent */
        }
        
        .task-entry .entry.sidechain {
            margin-left: 0; /* Override any margin for entries within task */
        }
        
        /* Ensure tool details in sub-agent conversations start collapsed */
        .task-entry .tool-details {
            display: none !important;
        }
        
        .task-entry .tool-call.expanded .tool-details {
            display: block !important;
        }
        
        /* Fix chevron rotation in sub-agent conversations */
        .task-entry .expanded .expand-icon {
            transform: rotate(90deg);
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
        }
        
        .role.user {
            background: #3498db;
            color: white;
        }
        
        .role.assistant {
            background: #27ae60;
            color: white;
        }
        
        .role.subagent {
            background: #f39c12;
            color: white;
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
        }
        
        .tool-header:hover {
            background: #e9ecef;
            margin: -10px;
            padding: 10px;
            border-radius: 4px;
        }
        
        .tool-name {
            font-weight: bold;
            color: #495057;
        }
        
        .tool-description {
            color: #6c757d;
            font-size: 0.9em;
        }
        
        .expand-icon {
            width: 20px;
            height: 20px;
            transition: transform 0.2s;
        }
        
        .expanded .expand-icon {
            transform: rotate(90deg);
        }
        
        .tool-details {
            display: none;
            margin-top: 10px;
            padding-top: 10px;
            border-top: 1px solid #dee2e6;
        }
        
        .expanded .tool-details {
            display: block;
        }
        
        .tool-input {
            background: #f1f3f5;
            padding: 10px;
            border-radius: 4px;
            overflow-x: auto;
            font-size: 0.85em;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            line-height: 1.4;
        }
        
        .tool-input pre {
            white-space: pre-wrap;
            word-wrap: break-word;
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
        .edit-diff {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 0;
            margin: 0;
            overflow: hidden;
        }
        
        .diff-header {
            background: #e9ecef;
            padding: 10px 15px;
            font-weight: bold;
            border-bottom: 1px solid #dee2e6;
        }
        
        .diff-header .file-path {
            color: #0056b3;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.9em;
        }
        
        .diff-header .replace-all {
            background: #6c757d;
            color: white;
            padding: 2px 8px;
            border-radius: 3px;
            font-size: 0.75em;
            margin-left: 10px;
        }
        
        .diff-content {
            background: #fafafa;
        }
        
        .diff-content.unified {
            padding: 0;
        }
        
        .diff-code {
            margin: 0;
            padding: 10px;
            overflow-x: auto;
            font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
            font-size: 0.85em;
            line-height: 1.4;
            white-space: pre;
            background: #fafafa;
        }
        
        .diff-line {
            display: block;
            padding: 0 15px;
            margin: 0;
            line-height: 1.4;
        }
        
        .diff-line.line-removed {
            background: #ffebee;
            color: #d32f2f;
        }
        
        .diff-line.line-added {
            background: #e8f5e9;
            color: #388e3c;
        }
        
        .diff-line.line-unchanged {
            background: transparent;
            color: #666;
        }
        
        .diff-code .line-number {
            color: #999;
            user-select: none;
            margin-right: 10px;
            display: inline-block;
            text-align: right;
            width: 30px;
        }
        
        .diff-code .line-prefix {
            font-weight: bold;
            margin-right: 8px;
        }
        
        /* Multi-edit styles */
        .multi-edit {
            background: #f8f9fa;
            border: 1px solid #dee2e6;
            border-radius: 4px;
            padding: 0;
            margin: 0;
            overflow: hidden;
        }
        
        .multi-edit .diff-header {
            margin-bottom: 0;
        }
        
        .multi-edit .edit-item {
            padding: 10px 15px;
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
    </style>
</head>
<body>
    <div class="container">
        <h1>Claude Code Conversation Log</h1>
        {{range .}}
            {{template "entry" .}}
        {{end}}
    </div>
    
    <script>
        // Use event delegation for tool call toggling
        document.addEventListener('click', (e) => {
            // Handle tool header clicks
            const toolHeader = e.target.closest('.tool-header');
            if (toolHeader) {
                e.preventDefault();
                e.stopPropagation();
                toolHeader.parentElement.classList.toggle('expanded');
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
        
    </script>
</body>
</html>

{{define "entry"}}
<div class="entry {{.Type}}{{if .IsSidechain}} sidechain{{end}}">
    <div class="entry-header">
        {{if .IsSidechain}}
            {{if eq .Role "user"}}
            <span class="role assistant">Assistant</span>
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
    
    <div class="content">{{.Content}}</div>
    
    {{if .ToolCalls}}
    <div class="tool-calls">
        {{range .ToolCalls}}
        <div class="tool-call">
            <div class="tool-header">
                <svg class="expand-icon" viewBox="0 0 20 20" fill="currentColor">
                    <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                </svg>
                <span class="tool-name">{{.Name}}</span>
                {{if .Description}}
                <span class="tool-description">{{.Description}}</span>
                {{end}}
            </div>
            <div class="tool-details">
                {{.Input}}
                {{if .TaskEntries}}
                <div style="margin-top: 15px;">
                    {{range .TaskEntries}}
                        <div class="task-entry">
                            {{template "entry" .}}
                        </div>
                    {{end}}
                </div>
                {{end}}
                {{if .Result}}
                <div class="tool-result-section" style="margin-top: 15px;">
                    <div class="result-header" style="cursor: pointer; user-select: none; display: flex; align-items: center; gap: 5px;">
                        <svg class="result-expand-icon" width="16" height="16" viewBox="0 0 20 20" fill="currentColor" style="transition: transform 0.2s;">
                            <path fill-rule="evenodd" d="M7.293 14.707a1 1 0 010-1.414L10.586 10 7.293 6.707a1 1 0 011.414-1.414l4 4a1 1 0 010 1.414l-4 4a1 1 0 01-1.414 0z" clip-rule="evenodd" />
                        </svg>
                        <strong>Result</strong>
                    </div>
                    <div class="result-content" style="display: none; margin-top: 10px;">
                        {{.Result.Content}}
                    </div>
                </div>
                {{end}}
            </div>
        </div>
        {{if .CompactView}}
        {{.CompactView}}
        {{end}}
        {{end}}
    </div>
    {{end}}
    
</div>
{{end}}`