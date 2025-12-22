// Internationalization module
const i18n = {
    en: {
        title: "Verdict Agent",
        placeholder: "Enter your idea or decision to make...",
        submit: "Get Verdict",
        loading: "Processing your request...",
        errorTitle: "Error",
        decisionTitle: "Decision",
        rulingLabel: "Ruling",
        rationaleLabel: "Rationale",
        rejectedLabel: "Rejected Options",
        todoTitle: "Execution Plan",
        decisionIdLabel: "Decision ID:",
        langToggle: "中文",
        charCount: "/10000",
        clarificationTitle: "Need More Information",
        clarificationSkip: "Skip & Decide",
        clarificationSubmit: "Submit Answers",
        answerPlaceholder: "Enter your answer...",
        stepClarify: "Analyzing context",
        stepSearch: "Searching for information",
        stepVerdict: "Making decision",
        stepPlan: "Generating execution plan"
    },
    zh: {
        title: "裁决代理",
        placeholder: "输入您的想法或需要做出的决定...",
        submit: "获取裁决",
        loading: "正在处理您的请求...",
        errorTitle: "错误",
        decisionTitle: "决策",
        rulingLabel: "裁决",
        rationaleLabel: "理由",
        rejectedLabel: "被否决的选项",
        todoTitle: "执行计划",
        decisionIdLabel: "决策ID:",
        langToggle: "English",
        charCount: "/10000",
        clarificationTitle: "需要更多信息",
        clarificationSkip: "跳过并决策",
        clarificationSubmit: "提交回答",
        answerPlaceholder: "请输入您的回答...",
        stepClarify: "分析上下文",
        stepSearch: "搜索相关信息",
        stepVerdict: "做出决策",
        stepPlan: "生成执行计划"
    }
};

let currentLang = 'en';

function toggleLanguage() {
    currentLang = currentLang === 'en' ? 'zh' : 'en';
    localStorage.setItem('verdict-agent-lang', currentLang);
    updateUILanguage();
}

function updateUILanguage() {
    const t = i18n[currentLang];

    // Update text content
    document.getElementById('title').textContent = t.title;
    document.getElementById('input').placeholder = t.placeholder;
    document.getElementById('submit-text').textContent = t.submit;
    document.getElementById('loading-text').textContent = t.loading;
    document.getElementById('error-title').textContent = t.errorTitle;
    document.getElementById('decision-title').textContent = t.decisionTitle;
    document.getElementById('ruling-label').textContent = t.rulingLabel;
    document.getElementById('rationale-label').textContent = t.rationaleLabel;
    document.getElementById('rejected-label').textContent = t.rejectedLabel;
    document.getElementById('todo-title').textContent = t.todoTitle;
    document.getElementById('decision-id-label').textContent = t.decisionIdLabel;
    document.getElementById('lang-text').textContent = t.langToggle;

    // Update clarification labels
    updateClarificationLabels();

    // Update document language
    document.documentElement.lang = currentLang;

    // Update page title
    document.title = t.title;
}

function updateClarificationLabels() {
    const t = i18n[currentLang];
    const clarificationTitle = document.getElementById('clarification-title');
    const skipText = document.getElementById('skip-text');
    const answerText = document.getElementById('answer-text');

    if (clarificationTitle) clarificationTitle.textContent = t.clarificationTitle;
    if (skipText) skipText.textContent = t.clarificationSkip;
    if (answerText) answerText.textContent = t.clarificationSubmit;

    // Update progress step labels
    const stepClarify = document.getElementById('step-clarify-text');
    const stepSearch = document.getElementById('step-search-text');
    const stepVerdict = document.getElementById('step-verdict-text');
    const stepPlan = document.getElementById('step-plan-text');

    if (stepClarify) stepClarify.textContent = t.stepClarify;
    if (stepSearch) stepSearch.textContent = t.stepSearch;
    if (stepVerdict) stepVerdict.textContent = t.stepVerdict;
    if (stepPlan) stepPlan.textContent = t.stepPlan;
}

function getCurrentLang() {
    return currentLang;
}

function initLanguage() {
    // Try to load from localStorage
    const savedLang = localStorage.getItem('verdict-agent-lang');
    if (savedLang && (savedLang === 'en' || savedLang === 'zh')) {
        currentLang = savedLang;
    } else {
        // Detect browser language
        const browserLang = navigator.language || navigator.userLanguage;
        if (browserLang.startsWith('zh')) {
            currentLang = 'zh';
        }
    }
    updateUILanguage();
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', initLanguage);
