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
    const clarification = document.getElementById('clarification');
    const clarificationReason = document.getElementById('clarification-reason');
    const questionsContainer = document.getElementById('questions-container');
    const results = document.getElementById('results');
    const rulingText = document.getElementById('ruling-text');
    const rationaleText = document.getElementById('rationale-text');
    const rejectedSection = document.getElementById('rejected-section');
    const rejectedList = document.getElementById('rejected-list');
    const todoContent = document.getElementById('todo-content');
    const decisionId = document.getElementById('decision-id');

    // State
    let currentInput = '';
    let currentQuestions = [];
    let progressTimer = null;
    let currentStep = 0;

    // Progress steps configuration
    const progressSteps = ['step-clarify', 'step-search', 'step-verdict', 'step-plan'];
    const stepDurations = [2000, 4000, 8000, 5000]; // Estimated time for each step

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

        currentInput = inputValue;
        showLoading(false);

        try {
            const response = await submitVerdict({ input: inputValue });
            handleResponse(response);
        } catch (err) {
            displayError(err.message);
        }
    };

    // Handle clarification form submission
    window.handleClarificationSubmit = async function(event) {
        event.preventDefault();

        const answers = collectAnswers();
        showLoading(true); // Skip clarify step since user already answered
        hideClarification();

        try {
            const response = await submitVerdict({
                input: currentInput,
                clarification: { answers: answers }
            });
            handleResponse(response);
        } catch (err) {
            displayError(err.message);
        }
    };

    // Skip clarification and get direct verdict
    window.skipClarification = async function() {
        showLoading(true); // Skip clarify step
        hideClarification();

        try {
            const response = await submitVerdict({
                input: currentInput,
                skip_clarify: true
            });
            handleResponse(response);
        } catch (err) {
            displayError(err.message);
        }
    };

    async function submitVerdict(payload) {
        const response = await fetch('/api/verdict', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        const data = await response.json();

        if (!response.ok) {
            throw new Error(data.error || 'An unknown error occurred');
        }

        return data;
    }

    function handleResponse(data) {
        hideLoading();

        if (data.status === 'clarification_needed') {
            displayClarification(data);
        } else {
            displayResults(data);
        }
    }

    function showLoading(skipClarifyStep) {
        submitBtn.disabled = true;
        loading.classList.remove('hidden');
        error.classList.add('hidden');
        results.classList.add('hidden');
        clarification.classList.add('hidden');

        // Reset progress steps
        resetProgressSteps();

        // Start progress animation
        currentStep = skipClarifyStep ? 1 : 0; // Skip clarify step if already answered
        startProgressAnimation();
    }

    function hideLoading() {
        submitBtn.disabled = false;

        // Complete all remaining steps
        completeAllSteps();

        // Stop progress timer
        if (progressTimer) {
            clearTimeout(progressTimer);
            progressTimer = null;
        }

        // Slight delay before hiding to show completion
        setTimeout(function() {
            loading.classList.add('hidden');
            resetProgressSteps();
        }, 500);
    }

    function resetProgressSteps() {
        progressSteps.forEach(function(stepId) {
            const step = document.getElementById(stepId);
            if (step) {
                step.classList.remove('active', 'completed');
            }
        });
        currentStep = 0;
    }

    function startProgressAnimation() {
        if (currentStep >= progressSteps.length) return;

        const stepId = progressSteps[currentStep];
        const step = document.getElementById(stepId);

        if (step) {
            step.classList.add('active');
        }

        // Schedule next step
        progressTimer = setTimeout(function() {
            if (step) {
                step.classList.remove('active');
                step.classList.add('completed');
            }
            currentStep++;
            startProgressAnimation();
        }, stepDurations[currentStep] || 3000);
    }

    function completeAllSteps() {
        progressSteps.forEach(function(stepId) {
            const step = document.getElementById(stepId);
            if (step) {
                step.classList.remove('active');
                step.classList.add('completed');
            }
        });
    }

    function hideClarification() {
        clarification.classList.add('hidden');
    }

    function displayError(message) {
        hideLoading();
        errorMessage.textContent = message;
        error.classList.remove('hidden');
        results.classList.add('hidden');
        clarification.classList.add('hidden');
    }

    function displayClarification(data) {
        hideLoading();
        error.classList.add('hidden');
        results.classList.add('hidden');
        clarification.classList.remove('hidden');

        // Show reason
        clarificationReason.textContent = data.reason || '';

        // Store questions
        currentQuestions = data.questions || [];

        // Build questions UI
        questionsContainer.innerHTML = '';
        currentQuestions.forEach(function(q, index) {
            const div = document.createElement('div');
            div.className = 'question-item';
            div.innerHTML = buildQuestionHTML(q, index);
            questionsContainer.appendChild(div);
        });

        // Update i18n labels
        if (typeof updateClarificationLabels === 'function') {
            updateClarificationLabels();
        }
    }

    function buildQuestionHTML(question, index) {
        const requiredMark = question.required ? '<span class="required">*</span>' : '';
        let inputHTML = '';

        if (question.type === 'choice' && question.options && question.options.length > 0) {
            inputHTML = '<div class="question-options">';
            question.options.forEach(function(opt, optIndex) {
                const optId = 'q_' + question.id + '_opt_' + optIndex;
                inputHTML += '<div class="option-item">' +
                    '<input type="radio" name="q_' + question.id + '" id="' + optId + '" value="' + escapeAttr(opt) + '"' +
                    (question.required ? ' required' : '') + '>' +
                    '<label for="' + optId + '">' + escapeHtml(opt) + '</label>' +
                    '</div>';
            });
            inputHTML += '</div>';
        } else if (question.type === 'multiple_choice' && question.options && question.options.length > 0) {
            inputHTML = '<div class="question-options">';
            question.options.forEach(function(opt, optIndex) {
                const optId = 'q_' + question.id + '_opt_' + optIndex;
                inputHTML += '<div class="option-item">' +
                    '<input type="checkbox" name="q_' + question.id + '" id="' + optId + '" value="' + escapeAttr(opt) + '">' +
                    '<label for="' + optId + '">' + escapeHtml(opt) + '</label>' +
                    '</div>';
            });
            inputHTML += '</div>';
        } else {
            inputHTML = '<input type="text" class="question-input" name="q_' + question.id + '" ' +
                'placeholder="' + (getCurrentLang() === 'zh' ? '请输入您的回答...' : 'Enter your answer...') + '"' +
                (question.required ? ' required' : '') + '>';
        }

        return '<label class="question-label">' +
            escapeHtml(question.question) + requiredMark +
            '</label>' + inputHTML;
    }

    function collectAnswers() {
        const answers = {};
        currentQuestions.forEach(function(q) {
            if (q.type === 'choice') {
                const selected = document.querySelector('input[name="q_' + q.id + '"]:checked');
                if (selected) {
                    answers[q.id] = selected.value;
                }
            } else if (q.type === 'multiple_choice') {
                const selected = document.querySelectorAll('input[name="q_' + q.id + '"]:checked');
                const values = [];
                selected.forEach(function(el) {
                    values.push(el.value);
                });
                if (values.length > 0) {
                    answers[q.id] = values.join(', ');
                }
            } else {
                const el = document.querySelector('input[name="q_' + q.id + '"]');
                if (el && el.value.trim()) {
                    answers[q.id] = el.value.trim();
                }
            }
        });
        return answers;
    }

    function displayResults(data) {
        hideLoading();
        error.classList.add('hidden');
        clarification.classList.add('hidden');
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

    function escapeAttr(text) {
        if (!text) return '';
        return text.replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    }

    // getCurrentLang() is provided by i18n.js

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
