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
        charCount: "/10000"
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
        charCount: "/10000"
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

    // Update document language
    document.documentElement.lang = currentLang;

    // Update page title
    document.title = t.title;
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
