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

    // Auth elements
    const authLoggedOut = document.getElementById('auth-logged-out');
    const authLoggedIn = document.getElementById('auth-logged-in');
    const userDisplay = document.getElementById('user-display');
    const authModal = document.getElementById('auth-modal');
    const authForm = document.getElementById('auth-form');
    const authModalTitle = document.getElementById('auth-modal-title');
    const authUsername = document.getElementById('auth-username');
    const authPassword = document.getElementById('auth-password');
    const authError = document.getElementById('auth-error');
    const authSubmitBtn = document.getElementById('auth-submit-btn');
    const authSwitchText = document.getElementById('auth-switch-text');
    const authSwitchLink = document.getElementById('auth-switch-link');
    const historyModal = document.getElementById('history-modal');
    const historyList = document.getElementById('history-list');
    const scoreCard = document.getElementById('score-card');
    const scoreFill = document.getElementById('score-fill');
    const scoreValue = document.getElementById('score-value');

    // State
    let currentInput = '';
    let currentQuestions = [];
    let progressTimer = null;
    let currentStep = 0;
    let authMode = 'login'; // 'login' or 'register'
    let authToken = localStorage.getItem('authToken');
    let currentUser = null;
    let currentHistoryId = null;
    let currentDoneCriteria = [];

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
        const headers = {
            'Content-Type': 'application/json'
        };
        if (authToken) {
            headers['Authorization'] = 'Bearer ' + authToken;
        }

        const response = await fetch('/api/verdict', {
            method: 'POST',
            headers: headers,
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

        // Display todo as rendered markdown with interactive checkboxes if logged in
        if (data.todo) {
            if (authToken && data.history_id) {
                currentHistoryId = data.history_id;
                // Extract done criteria from API response
                if (data.done_criteria && data.done_criteria.length > 0) {
                    currentDoneCriteria = data.done_criteria.map(function(text, index) {
                        return { index: index, text: text, completed: false };
                    });
                }
                todoContent.innerHTML = renderMarkdownWithCheckboxes(data.todo);
                setupCheckboxListeners();
                showScoreCard();
                updateScore(0);
            } else {
                todoContent.innerHTML = renderMarkdown(data.todo);
                hideScoreCard();
            }
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

    function renderMarkdownWithCheckboxes(markdown) {
        if (!markdown) return '';

        // Split by lines to handle Done Criteria section specially
        const lines = markdown.split('\n');
        let inDoneCriteria = false;
        let doneCriteriaIndex = 0;
        let processedLines = [];

        for (let i = 0; i < lines.length; i++) {
            let line = lines[i];

            // Check for Done Criteria header
            if (line.match(/^## Done Criteria/i)) {
                inDoneCriteria = true;
                processedLines.push(line);
                continue;
            }

            // Check for next section header (exit Done Criteria)
            if (inDoneCriteria && line.match(/^## /)) {
                inDoneCriteria = false;
            }

            // Convert Done Criteria bullet items to checkboxes
            if (inDoneCriteria && line.match(/^- [^\[\]]/)) {
                const text = line.substring(2); // Remove "- "
                const completed = currentDoneCriteria[doneCriteriaIndex] && currentDoneCriteria[doneCriteriaIndex].completed;
                const checkedAttr = completed ? ' checked' : '';
                const completedClass = completed ? ' class="completed"' : '';
                line = '- [DONE_CRITERIA:' + doneCriteriaIndex + ':' + checkedAttr + '] ' + text;
                doneCriteriaIndex++;
            }

            processedLines.push(line);
        }

        let html = escapeHtml(processedLines.join('\n'));

        // Headers
        html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
        html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

        // Phase task checkboxes (visual only, not tracked for score)
        html = html.replace(/^- \[ \] (.+)$/gm, function(match, text) {
            return '<li class="phase-task"><input type="checkbox" class="phase-checkbox"> ' + text + '</li>';
        });
        html = html.replace(/^- \[x\] (.+)$/gm, function(match, text) {
            return '<li class="phase-task completed"><input type="checkbox" class="phase-checkbox" checked> ' + text + '</li>';
        });

        // Done Criteria checkboxes (tracked for score)
        html = html.replace(/- \[DONE_CRITERIA:(\d+):([^\]]*)\] (.+)/g, function(match, idx, checked, text) {
            const isChecked = checked.includes('checked');
            const checkedAttr = isChecked ? ' checked' : '';
            const completedClass = isChecked ? ' completed' : '';
            return '<li class="done-criteria-item' + completedClass + '" data-criteria-index="' + idx + '"><input type="checkbox" class="done-criteria-checkbox" data-criteria-index="' + idx + '"' + checkedAttr + '> ' + text + '</li>';
        });

        // Regular bullet items
        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');

        // Wrap consecutive list items in ul
        html = html.replace(/(<li[^>]*>.*<\/li>\n?)+/g, function(match) {
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

        if (!html.startsWith('<')) {
            html = '<p>' + html + '</p>';
        }

        return html;
    }

    function setupCheckboxListeners() {
        // Done Criteria checkboxes - tracked for score
        const doneCriteriaCheckboxes = todoContent.querySelectorAll('.done-criteria-checkbox');
        doneCriteriaCheckboxes.forEach(function(checkbox) {
            checkbox.addEventListener('change', handleDoneCriteriaChange);
        });

        // Phase task checkboxes - visual only (strikethrough)
        const phaseCheckboxes = todoContent.querySelectorAll('.phase-checkbox');
        phaseCheckboxes.forEach(function(checkbox) {
            checkbox.addEventListener('change', handlePhaseCheckboxChange);
        });
    }

    function handlePhaseCheckboxChange(event) {
        const checkbox = event.target;
        const li = checkbox.closest('li');

        if (checkbox.checked) {
            li.classList.add('completed');
        } else {
            li.classList.remove('completed');
        }
    }

    async function handleDoneCriteriaChange(event) {
        const checkbox = event.target;
        const index = parseInt(checkbox.dataset.criteriaIndex, 10);
        const li = checkbox.closest('li');

        if (checkbox.checked) {
            li.classList.add('completed');
        } else {
            li.classList.remove('completed');
        }

        // Update local state
        if (currentDoneCriteria[index]) {
            currentDoneCriteria[index].completed = checkbox.checked;
        }

        // Calculate and display score
        const completed = currentDoneCriteria.filter(function(c) { return c.completed; }).length;
        const total = currentDoneCriteria.length;
        const score = total > 0 ? Math.round((completed / total) * 100) : 0;
        updateScore(score);

        // Save to backend if we have a history ID
        if (currentHistoryId && authToken) {
            try {
                await updateDoneCriteria(currentHistoryId, currentDoneCriteria);
            } catch (err) {
                console.error('Failed to save progress:', err);
            }
        }
    }

    async function updateDoneCriteria(historyId, criteria) {
        const response = await fetch('/api/history/' + historyId + '/criteria', {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + authToken
            },
            body: JSON.stringify({ done_criteria: criteria })
        });

        if (!response.ok) {
            throw new Error('Failed to update criteria');
        }

        return response.json();
    }

    function showScoreCard() {
        if (scoreCard) scoreCard.classList.remove('hidden');
    }

    function hideScoreCard() {
        if (scoreCard) scoreCard.classList.add('hidden');
    }

    function updateScore(score) {
        if (scoreFill) scoreFill.style.width = score + '%';
        if (scoreValue) scoreValue.textContent = score + '%';
    }

    // === Auth Functions ===

    // Initialize auth state on page load
    if (authToken) {
        checkAuthStatus();
    }

    async function checkAuthStatus() {
        try {
            const response = await fetch('/api/auth/me', {
                headers: { 'Authorization': 'Bearer ' + authToken }
            });
            if (response.ok) {
                const data = await response.json();
                currentUser = data;
                showLoggedInState();
            } else {
                clearAuth();
            }
        } catch (err) {
            clearAuth();
        }
    }

    function showLoggedInState() {
        if (authLoggedOut) authLoggedOut.classList.add('hidden');
        if (authLoggedIn) authLoggedIn.classList.remove('hidden');
        if (userDisplay && currentUser) userDisplay.textContent = currentUser.username;
    }

    function showLoggedOutState() {
        if (authLoggedIn) authLoggedIn.classList.add('hidden');
        if (authLoggedOut) authLoggedOut.classList.remove('hidden');
    }

    function clearAuth() {
        authToken = null;
        currentUser = null;
        localStorage.removeItem('authToken');
        showLoggedOutState();
    }

    window.showAuthModal = function(mode) {
        authMode = mode;
        updateAuthModal();
        if (authModal) authModal.classList.remove('hidden');
        if (authUsername) authUsername.focus();
    };

    window.hideAuthModal = function() {
        if (authModal) authModal.classList.add('hidden');
        if (authForm) authForm.reset();
        if (authError) authError.classList.add('hidden');
    };

    window.switchAuthMode = function(event) {
        event.preventDefault();
        authMode = authMode === 'login' ? 'register' : 'login';
        updateAuthModal();
    };

    function updateAuthModal() {
        const isLogin = authMode === 'login';
        const lang = getCurrentLang();

        if (authModalTitle) {
            authModalTitle.textContent = isLogin
                ? (lang === 'zh' ? '登录' : 'Login')
                : (lang === 'zh' ? '注册' : 'Register');
        }
        if (authSubmitBtn) {
            authSubmitBtn.textContent = isLogin
                ? (lang === 'zh' ? '登录' : 'Login')
                : (lang === 'zh' ? '注册' : 'Register');
        }
        if (authSwitchText) {
            authSwitchText.textContent = isLogin
                ? (lang === 'zh' ? '还没有账户？' : "Don't have an account?")
                : (lang === 'zh' ? '已有账户？' : 'Already have an account?');
        }
        if (authSwitchLink) {
            authSwitchLink.textContent = isLogin
                ? (lang === 'zh' ? '注册' : 'Register')
                : (lang === 'zh' ? '登录' : 'Login');
        }
    }

    window.handleAuth = async function(event) {
        event.preventDefault();
        if (authError) authError.classList.add('hidden');

        const username = authUsername ? authUsername.value.trim() : '';
        const password = authPassword ? authPassword.value : '';

        if (!username || !password) return;

        try {
            const endpoint = authMode === 'login' ? '/api/auth/login' : '/api/auth/register';
            const response = await fetch(endpoint, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username: username, password: password })
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Authentication failed');
            }

            authToken = data.token;
            currentUser = { id: data.user_id, username: data.username };
            localStorage.setItem('authToken', authToken);
            showLoggedInState();
            hideAuthModal();
        } catch (err) {
            if (authError) {
                authError.textContent = err.message;
                authError.classList.remove('hidden');
            }
        }
    };

    window.logout = async function() {
        try {
            await fetch('/api/auth/logout', {
                method: 'POST',
                headers: { 'Authorization': 'Bearer ' + authToken }
            });
        } catch (err) {
            // Ignore errors
        }
        clearAuth();
        hideScoreCard();
    };

    // === History Functions ===

    window.showHistory = async function() {
        if (!authToken) return;

        if (historyModal) historyModal.classList.remove('hidden');

        try {
            const response = await fetch('/api/history', {
                headers: { 'Authorization': 'Bearer ' + authToken }
            });

            if (!response.ok) throw new Error('Failed to load history');

            const history = await response.json();
            displayHistory(history);
        } catch (err) {
            if (historyList) {
                historyList.innerHTML = '<div class="history-empty">' +
                    (getCurrentLang() === 'zh' ? '加载历史记录失败' : 'Failed to load history') +
                    '</div>';
            }
        }
    };

    window.hideHistoryModal = function() {
        if (historyModal) historyModal.classList.add('hidden');
    };

    function displayHistory(history) {
        if (!historyList) return;

        if (!history || history.length === 0) {
            historyList.innerHTML = '<div class="history-empty">' +
                (getCurrentLang() === 'zh' ? '暂无决策历史' : 'No decision history yet') +
                '</div>';
            return;
        }

        historyList.innerHTML = history.map(function(item) {
            const date = new Date(item.created_at).toLocaleDateString();
            const inputPreview = item.input.length > 60
                ? item.input.substring(0, 60) + '...'
                : item.input;

            return '<div class="history-item" onclick="loadHistoryItem(\'' + item.id + '\')">' +
                '<div class="history-item-header">' +
                '<span class="history-item-input">' + escapeHtml(inputPreview) + '</span>' +
                '<span class="history-item-score">' + Math.round(item.score) + '%</span>' +
                '</div>' +
                '<div class="history-item-date">' + date + '</div>' +
                '</div>';
        }).join('');
    }

    window.loadHistoryItem = async function(historyId) {
        hideHistoryModal();

        try {
            const response = await fetch('/api/history/' + historyId, {
                headers: { 'Authorization': 'Bearer ' + authToken }
            });

            if (!response.ok) throw new Error('Failed to load history item');

            const item = await response.json();

            // Display the saved decision
            currentHistoryId = item.id;
            currentDoneCriteria = item.done_criteria || [];

            // Parse and display the verdict
            let decision;
            try {
                decision = typeof item.verdict === 'string'
                    ? JSON.parse(item.verdict)
                    : item.verdict;
            } catch (e) {
                decision = item.verdict;
            }

            if (decision && decision.verdict) {
                rulingText.textContent = decision.verdict.ruling || '';
                rationaleText.textContent = decision.verdict.rationale || '';

                if (decision.verdict.rejected && decision.verdict.rejected.length > 0) {
                    rejectedSection.classList.remove('hidden');
                    rejectedList.innerHTML = '';
                    decision.verdict.rejected.forEach(function(r) {
                        const li = document.createElement('li');
                        li.innerHTML = '<div class="rejected-option">' + escapeHtml(r.option) + '</div>' +
                                       '<div class="rejected-reason">' + escapeHtml(r.reason) + '</div>';
                        rejectedList.appendChild(li);
                    });
                } else {
                    rejectedSection.classList.add('hidden');
                }
            }

            // Display todo with saved checkbox states
            if (item.todo) {
                todoContent.innerHTML = renderMarkdownWithSavedStates(item.todo, currentDoneCriteria);
                setupCheckboxListeners();
                showScoreCard();
                updateScore(Math.round(item.score));
            }

            decisionId.textContent = item.decision_id;

            // Show results
            error.classList.add('hidden');
            clarification.classList.add('hidden');
            results.classList.remove('hidden');

        } catch (err) {
            displayError(err.message);
        }
    };

    function renderMarkdownWithSavedStates(markdown, criteria) {
        if (!markdown) return '';

        let html = escapeHtml(markdown);

        // Headers
        html = html.replace(/^### (.+)$/gm, '<h3>$1</h3>');
        html = html.replace(/^## (.+)$/gm, '<h2>$1</h2>');
        html = html.replace(/^# (.+)$/gm, '<h1>$1</h1>');

        // Checkboxes - restore saved states
        let checkboxIndex = 0;
        html = html.replace(/^- \[[ x]\] (.+)$/gm, function(match, text) {
            const idx = checkboxIndex++;
            const criterion = criteria.find(function(c) { return c.index === idx; });
            const checked = criterion && criterion.completed;
            const checkedAttr = checked ? ' checked' : '';
            const completedClass = checked ? ' class="completed"' : '';
            return '<li data-index="' + idx + '"' + completedClass + '><input type="checkbox"' + checkedAttr + ' data-criteria-index="' + idx + '"> ' + text + '</li>';
        });
        html = html.replace(/^- (.+)$/gm, '<li>$1</li>');

        // Wrap consecutive list items in ul
        html = html.replace(/(<li[^>]*>.*<\/li>\n?)+/g, function(match) {
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

        if (!html.startsWith('<')) {
            html = '<p>' + html + '</p>';
        }

        return html;
    }
})();
