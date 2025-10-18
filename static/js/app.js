// 智能食谱助手 - 前端交互逻辑
class RecipeAssistant {
    constructor() {
        this.ingredients = [];
        this.initializeElements();
        this.bindEvents();
        this.hideAllContainers();
    }

    // 初始化DOM元素
    initializeElements() {
        // 表单元素
        this.ingredientInput = document.getElementById('ingredientInput');
        this.ingredientTags = document.getElementById('ingredientTags');
        this.dishNameInput = document.getElementById('dishName');

        // 按钮
        this.searchBtn = document.getElementById('searchBtn');
        this.clearBtn = document.getElementById('clearBtn');

        // 容器
        this.loadingContainer = document.getElementById('loadingContainer');
        this.resultContainer = document.getElementById('resultContainer');
        this.errorContainer = document.getElementById('errorContainer');

        // 结果相关
        this.resultContent = document.getElementById('resultContent');
        this.supplementaryInfo = document.getElementById('supplementaryInfo');
        this.supplementaryContent = document.getElementById('supplementaryContent');
        this.serveCount = document.getElementById('serveCount');
        this.timestamp = document.getElementById('timestamp');
        this.errorMessage = document.getElementById('errorMessage');

        // 标签页
        this.searchTabs = document.getElementById('searchTabs');
        this.ingredientsTab = document.getElementById('ingredients-tab');
        this.dishTab = document.getElementById('dish-tab');

        // 模态框
        this.errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
        this.modalErrorMessage = document.getElementById('modalErrorMessage');
    }

    // 绑定事件
    bindEvents() {
        // 食材输入事件
        this.ingredientInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.addIngredient();
            }
        });

        // 搜索按钮事件
        this.searchBtn.addEventListener('click', () => {
            this.performSearch();
        });

        // 清空按钮事件
        this.clearBtn.addEventListener('click', () => {
            this.clearAll();
        });

        // 标签页切换事件
        this.ingredientsTab.addEventListener('shown.bs.tab', () => {
            this.ingredientInput.focus();
        });

        this.dishTab.addEventListener('shown.bs.tab', () => {
            this.dishNameInput.focus();
        });

        // 表单回车事件
        this.dishNameInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.performSearch();
            }
        });

        // 页面加载完成事件
        document.addEventListener('DOMContentLoaded', () => {
            this.ingredientInput.focus();
        });
    }

    // 添加食材标签
    addIngredient() {
        const input = this.ingredientInput.value.trim();
        if (input === '') {
            return;
        }

        // 检查重复
        if (this.ingredients.includes(input)) {
            this.showError('该食材已添加，请输入其他食材');
            return;
        }

        this.ingredients.push(input);
        this.renderIngredientTags();
        this.ingredientInput.value = '';
        this.ingredientInput.focus();
    }

    // 移除食材标签
    removeIngredient(index) {
        this.ingredients.splice(index, 1);
        this.renderIngredientTags();
    }

    // 渲染食材标签
    renderIngredientTags() {
        this.ingredientTags.innerHTML = '';

        if (this.ingredients.length === 0) {
            this.ingredientTags.innerHTML = '<span class="text-muted">暂无食材，请在上方输入食材并按回车添加</span>';
            return;
        }

        this.ingredients.forEach((ingredient, index) => {
            const tag = document.createElement('span');
            tag.className = 'ingredient-tag';
            tag.innerHTML = `
                ${ingredient}
                <span class="remove-tag" onclick="recipeAssistant.removeIngredient(${index})">×</span>
            `;
            this.ingredientTags.appendChild(tag);
        });
    }

    // 执行搜索
    async performSearch() {
        const activeTab = document.querySelector('.nav-link.active').id;
        let requestData = {};

        if (activeTab === 'ingredients-tab') {
            if (this.ingredients.length === 0) {
                this.showError('请添加至少一个食材');
                return;
            }
            requestData = {
                queryType: 'ingredients',
                ingredients: this.ingredients
            };
        } else {
            const dishName = this.dishNameInput.value.trim();
            if (dishName === '') {
                this.showError('请输入菜品名称');
                return;
            }
            requestData = {
                queryType: 'dish',
                dishName: dishName
            };
        }

        this.showLoading();

        try {
            const response = await fetch('/api/recipes', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(requestData)
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.message || '请求失败');
            }

            this.showResult(data);
        } catch (error) {
            console.error('搜索错误:', error);
            this.showError(error.message || '搜索过程中出现错误，请稍后重试');
        } finally {
            this.hideLoading();
        }
    }

    // 显示加载状态
    showLoading() {
        this.hideAllContainers();
        this.searchBtn.disabled = true;
        this.clearBtn.disabled = true;
        this.loadingContainer.style.display = 'block';

        // 添加加载动画文本变化
        this.animateLoadingText();
    }

    // 动画加载文本
    animateLoadingText() {
        const loadingText = document.querySelector('.loading-text');
        const messages = [
            'AI正在精心分析，请稍候...',
            '正在匹配最佳食谱方案...',
            '计算营养成分和制作要点...',
            '整理烹饪技巧和注意事项...'
        ];
        let index = 0;

        this.loadingInterval = setInterval(() => {
            loadingText.textContent = messages[index];
            index = (index + 1) % messages.length;
        }, 2000);
    }

    // 隐藏加载状态
    hideLoading() {
        if (this.loadingInterval) {
            clearInterval(this.loadingInterval);
        }
        this.loadingContainer.style.display = 'none';
        this.searchBtn.disabled = false;
        this.clearBtn.disabled = false;
    }

    // 显示结果
    showResult(data) {
        this.hideAllContainers();
        this.resultContainer.style.display = 'block';

        // 设置元数据
        this.serveCount.textContent = this.getQueryTypeLabel(data.type);
        this.timestamp.textContent = this.formatTimestamp(data.timestamp);

        // 处理结果内容
        const formattedContent = this.formatContent(data.result);
        this.resultContent.innerHTML = formattedContent;

        // 显示附加信息
        if (data.supplementaryData && Object.keys(data.supplementaryData).length > 0) {
            this.showSupplementaryData(data.supplementaryData);
        }

        // 添加动画效果
        this.resultContainer.classList.add('fade-in');

        // 滚动到结果区域
        this.scrollToResult();
    }

    // 显示附加信息
    showSupplementaryData(data) {
        let content = '<div class="row">';

        if (data.ai_available) {
            content += `
                <div class="col-md-6 mb-3">
                    <div class="card border-success">
                        <div class="card-body">
                            <h6 class="card-title text-success">
                                <i class="bi bi-robot"></i> AI分析
                            </h6>
                            <p class="card-text small">智能食谱分析和个性化推荐已启用</p>
                        </div>
                    </div>
                </div>
            `;
        }

        if (data.api_available) {
            content += `
                <div class="col-md-6 mb-3">
                    <div class="card border-info">
                        <div class="card-body">
                            <h6 class="card-title text-info">
                                <i class="bi bi-database"></i> 数据源
                            </h6>
                            <p class="card-text small">${data.nutrition_tips || '专业食谱数据支持'}</p>
                        </div>
                    </div>
                </div>
            `;
        }

        if (data.api_recipes && data.api_recipes.length > 0) {
            content += `
                <div class="col-12">
                    <h6 class="text-muted mb-3">
                        <i class="bi bi-book"></i> 参考食谱 (${data.api_recipes.length}个)
                    </h6>
                    <div class="row">
            `;

            data.api_recipes.forEach((recipe, index) => {
                content += `
                    <div class="col-md-4 mb-3">
                        <div class="card">
                            <div class="card-body">
                                <h6 class="card-title">${recipe.title}</h6>
                                ${recipe.readyInMinutes ? `<p class="card-text small"><i class="bi bi-clock"></i> ${recipe.readyInMinutes}分钟</p>` : ''}
                                ${recipe.servings ? `<p class="card-text small"><i class="bi bi-people"></i> ${recipe.servings}人份</p>` : ''}
                            </div>
                        </div>
                    </div>
                `;
            });

            content += '</div></div>';
        }

        content += '</div>';

        this.supplementaryContent.innerHTML = content;
        this.supplementaryInfo.style.display = 'block';
    }

    // 格式化内容
    formatContent(content) {
        if (!content) return '<p>暂无内容</p>';

        // 处理Markdown格式
        let formatted = content
            // 处理标题
            .replace(/^### (.+)$/gm, '<h3>$1</h3>')
            .replace(/^## (.+)$/gm, '<h2>$1</h2>')
            .replace(/^# (.+)$/gm, '<h1>$1</h1>')
            // 处理粗体
            .replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>')
            // 处理斜体
            .replace(/\*(.+?)\*/g, '<em>$1</em>')
            // 处理列表
            .replace(/^- (.+)$/gm, '<li>$1</li>')
            .replace(/^\d+\. (.+)$/gm, '<li>$1</li>')
            // 处理换行
            .replace(/\n\n/g, '</p><p>')
            .replace(/\n/g, '<br>');

        // 包装列表
        formatted = formatted.replace(/(<li>.*?<\/li>)/gs, (match) => {
            if (!match.includes('<ul>') && !match.includes('<ol>')) {
                if (/^\d+/.test(match)) {
                    return `<ol>${match}</ol>`;
                } else {
                    return `<ul>${match}</ul>`;
                }
            }
            return match;
        });

        // 添加段落标签
        if (!formatted.startsWith('<h') && !formatted.startsWith('<')) {
            formatted = `<p>${formatted}</p>`;
        }

        return formatted;
    }

    // 显示错误
    showError(message) {
        this.hideAllContainers();
        this.errorContainer.style.display = 'block';
        this.errorMessage.textContent = message;

        // 同时显示模态框
        this.modalErrorMessage.textContent = message;
        this.errorModal.show();
    }

    // 隐藏所有容器
    hideAllContainers() {
        this.loadingContainer.style.display = 'none';
        this.resultContainer.style.display = 'none';
        this.errorContainer.style.display = 'none';
    }

    // 清空所有内容
    clearAll() {
        this.ingredients = [];
        this.renderIngredientTags();
        this.dishNameInput.value = '';
        this.ingredientInput.value = '';
        this.hideAllContainers();

        // 聚焦到当前激活的输入框
        const activeTab = document.querySelector('.nav-link.active').id;
        if (activeTab === 'ingredients-tab') {
            this.ingredientInput.focus();
        } else {
            this.dishNameInput.focus();
        }
    }

    // 获取查询类型标签
    getQueryTypeLabel(type) {
        return type === 'ingredients' ? '食材分析' : '菜品制作';
    }

    // 格式化时间戳
    formatTimestamp(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    }

    // 滚动到结果区域
    scrollToResult() {
        setTimeout(() => {
            this.resultContainer.scrollIntoView({
                behavior: 'smooth',
                block: 'start'
            });
        }, 100);
    }

    // 添加键盘快捷键支持
    handleKeyboardShortcuts(event) {
        // Ctrl/Cmd + Enter 执行搜索
        if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
            event.preventDefault();
            this.performSearch();
        }

        // Escape 清空内容
        if (event.key === 'Escape') {
            this.clearAll();
        }
    }
}

// 初始化应用
const recipeAssistant = new RecipeAssistant();

// 添加键盘快捷键支持
document.addEventListener('keydown', (event) => {
    recipeAssistant.handleKeyboardShortcuts(event);
});

// 添加页面可见性变化监听
document.addEventListener('visibilitychange', () => {
    if (!document.hidden && recipeAssistant.loadingContainer.style.display !== 'block') {
        // 页面重新可见时，聚焦到当前激活的输入框
        const activeTab = document.querySelector('.nav-link.active').id;
        if (activeTab === 'ingredients-tab') {
            recipeAssistant.ingredientInput.focus();
        } else {
            recipeAssistant.dishNameInput.focus();
        }
    }
});

// 添加错误处理
window.addEventListener('error', (event) => {
    console.error('页面错误:', event.error);
});

// 全局函数（供HTML内联事件使用）
window.recipeAssistant = recipeAssistant;