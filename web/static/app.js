// Main application logic
(function() {
    'use strict';

    // DOM Elements
    const form = document.getElementById('verdict-form');
    const input = document.getElementById('input');
    const submitBtn = document.getElementById('submit-btn');
    const charCount = document.getElementById('char-count');
    const loading = document.getElementById('loading');
    const error = document.getElementById('error');
    const errorMessage = document.getElementById('error-message');
    const results = document.getElementById('results');
    const rulingText = document.getElementById('ruling-text');
    const rationaleText = document.getElementById('rationale-text');
    const rejectedSection = document.getElementById('rejected-section');
    const rejectedList = document.getElementById('rejected-list');
    const todoContent = document.getElementById('todo-content');
    const decisionId = document.getElementById('decision-id');

    // Update character count
    input.addEventListener('input', function() {
        charCount.textContent = input.value.length;
    });

    // Handle form submission
    window.handleSubmit = async function(event) {
        event.preventDefault();

        const inputValue = input.value.trim();
        if (!inputValue) {
            return;
        }

        // Show loading, hide others
        showLoading();

        try {
            const response = await submitVerdict(inputValue);
            displayResults(response);
        } catch (err) {
            displayError(err.message);
        }
    };

    async function submitVerdict(inputText) {
        const response = await fetch('/api/verdict', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ input: inputText })
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'An unknown error occurred');
        }

        return data;
    }

    function showLoading() {
        submitBtn.disabled = true;
        loading.classList.remove('hidden');
        error.classList.add('hidden');
        results.classList.add('hidden');
    }

    function hideLoading() {
        submitBtn.disabled = false;
        loading.classList.add('hidden');
    }

    function displayError(message) {
        hideLoading();
        errorMessage.textContent = message;
        error.classList.remove('hidden');
        results.classList.add('hidden');
    }

    function displayResults(data) {
        hideLoading();
        error.classList.add('hidden');
        results.classList.remove('hidden');

        // Parse decision JSON
        let decision;
        try {
            decision = typeof data.decision === 'string'
                ? JSON.parse(data.decision)
                : data.decision;
        } catch (e) {
            decision = data.decision;
        }

        // Display verdict
        if (decision && decision.verdict) {
            rulingText.textContent = decision.verdict.ruling || '';
            rationaleText.textContent = decision.verdict.rationale || '';

            // Display rejected options
            if (decision.verdict.rejected && decision.verdict.rejected.length > 0) {
                rejectedSection.classList.remove('hidden');
                rejectedList.innerHTML = '';
                decision.verdict.rejected.forEach(function(item) {
                    const li = document.createElement('li');
                    li.innerHTML = '<div class="rejected-option">' + escapeHtml(item.option) + '</div>' +
                                   '<div class="rejected-reason">' + escapeHtml(item.reason) + '</div>';
                    rejectedList.appendChild(li);
                });
            } else {
                rejectedSection.classList.add('hidden');
            }
        }

        // Display todo as rendered markdown
        if (data.todo) {
            todoContent.innerHTML = renderMarkdown(data.todo);
        }

        // Display decision ID
        if (data.decision_id) {
            decisionId.textContent = data.decision_id;
        }
    }

    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function renderMarkdown(markdown) {
        if (!markdown) return '';

        // Simple markdown renderer
        let html = escapeHtml(markdown);

        // Headers
        html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
        html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

        // Checkboxes
        html = html.replace(/^- \[ \] (.+)$/gm, '<li><input type="checkbox" disabled> $1</li>');
        html = html.replace(/^- \[x\] (.+)$/gm, '<li><input type="checkbox" checked disabled> $1</li>');
        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');

        // Wrap consecutive list items in ul
        html = html.replace(/(<li>.*<\/li>\n?)+/g, function(match) {
            return '<ul>' + match + '</ul>';
        });

        // Bold
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');

        // Italic
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');

        // Code blocks
        html = html.replace(/`(.+?)`/g, '<code>$1</code>');

        // Line breaks
        html = html.replace(/\n\n/g, '</p><p>');
        html = html.replace(/\n/g, '<br>');

        // Wrap in paragraphs if not already wrapped
        if (!html.startsWith('<')) {
            html = '<p>' + html + '</p>';
        }

        return html;
    }
})();
