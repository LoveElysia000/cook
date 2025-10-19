// 智能食谱助手 - 前端交互逻辑
class RecipeAssistant {
    constructor() {
        this.ingredients = [];
        this.currentSearchType = 'ingredients'; // ingredients 或 dish
        this.initializeElements();
        this.bindEvents();
        this.hideAllContainers();
        this.initializeHints();
    }

    // 初始化DOM元素
    initializeElements() {
        // 智能搜索元素
        this.smartSearchInput = document.getElementById('smartSearchInput');
        this.typeBadge = document.getElementById('typeBadge');
        this.autocompleteDropdown = document.getElementById('autocompleteDropdown');
        this.autocompleteItems = document.getElementById('autocompleteItems');
        this.ingredientsDisplay = document.getElementById('ingredientsDisplay');
        this.ingredientTagsCompact = document.getElementById('ingredientTagsCompact');

        // 按钮
        this.searchBtn = document.getElementById('searchBtn');
        this.clearBtn = document.getElementById('clearBtn');
        this.voiceSearchBtn = document.getElementById('voiceSearchBtn');
        this.menuBtn = document.getElementById('menuBtn');
        this.quickAnalysisBtn = document.getElementById('quickAnalysisBtn');
        this.clearResultsBtn = document.getElementById('clearResultsBtn');
        this.closeDetailsBtn = document.getElementById('closeDetailsBtn');

        // 容器
        this.loadingContainer = document.getElementById('loadingContainer');
        this.resultsSection = document.getElementById('resultsSection');
        this.errorContainer = document.getElementById('errorContainer');
        this.resultDetails = document.getElementById('resultDetails');
        this.recipeCardsContainer = document.getElementById('recipeCardsContainer');

        // 结果相关
        this.serveCount = document.getElementById('serveCount');
        this.timestamp = document.getElementById('timestamp');
        this.errorMessage = document.getElementById('errorMessage');
        this.detailsContent = document.getElementById('detailsContent');

        // 提示相关
        this.hintContent = document.getElementById('hintContent');
        this.hintTabs = document.querySelectorAll('.hint-tab');

        // 附加信息
        this.supplementaryInfo = document.getElementById('supplementaryInfo');
        this.supplementaryContent = document.getElementById('supplementaryContent');

        // 模态框
        this.errorModal = new bootstrap.Modal(document.getElementById('errorModal'));
        this.modalErrorMessage = document.getElementById('modalErrorMessage');

        // 加载动画元素
        this.chefLoader = document.getElementById('chefLoader');
        this.spoonLoader = document.getElementById('spoonLoader');
        this.dynamicLoadingText = document.getElementById('dynamicLoadingText');
        this.analysisSteps = document.getElementById('analysisSteps');
        this.analysisStepElements = this.analysisSteps.querySelectorAll('.analysis-step');
    }

    // 绑定事件
    bindEvents() {
        // 智能搜索输入事件
        this.smartSearchInput.addEventListener('input', (e) => {
            this.handleSearchInput(e.target.value);
        });

        this.smartSearchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                this.handleSearchSubmit();
            }
        });

        this.smartSearchInput.addEventListener('focus', () => {
            this.showAutocomplete();
        });

        // 点击外部关闭自动补全
        document.addEventListener('click', (e) => {
            if (!e.target.closest('.search-input-wrapper') && !e.target.closest('.autocomplete-dropdown')) {
                this.hideAutocomplete();
            }
        });

        // 搜索按钮事件
        this.searchBtn.addEventListener('click', () => {
            this.performSearch();
        });

        // 清空按钮事件
        this.clearBtn.addEventListener('click', () => {
            this.clearSearchInput();
        });

        this.clearResultsBtn.addEventListener('click', () => {
            this.clearResults();
        });

        // 语音搜索按钮事件
        this.voiceSearchBtn.addEventListener('click', () => {
            this.startVoiceSearch();
        });

        // 顶部按钮事件
        this.menuBtn.addEventListener('click', () => {
            this.showMenu();
        });

        this.quickAnalysisBtn.addEventListener('click', () => {
            this.quickAnalysis();
        });

        // 关闭详情按钮事件
        this.closeDetailsBtn.addEventListener('click', () => {
            this.hideDetails();
        });

        // 提示标签切换事件
        this.hintTabs.forEach(tab => {
            tab.addEventListener('click', (e) => {
                this.switchHintTab(e.target.closest('.hint-tab'));
            });
        });

        // 页面加载完成事件
        document.addEventListener('DOMContentLoaded', () => {
            this.smartSearchInput.focus();
        });
    }

    // 初始化提示内容
    initializeHints() {
        this.hintsData = {
            ingredients: [
                '🥚 鸡蛋、鸭蛋', '🥬 西红柿、青菜', '🥩 猪肉、鸡肉', '🥕 土豆、萝卜',
                '🫛 豆类', '🍄 蘑菇', '🧀 奶酪', '🥦 西兰花'
            ],
            dishes: [
                '🥘 家常菜：红烧肉、糖醋排骨', '🍛 地方菜：宫保鸡丁、水煮鱼',
                '🥣 汤类：西红柿鸡蛋汤', '🍜 面食：炸酱面、牛肉面'
            ],
            seasonal: [
                '🌸 春季：春笋、韭菜、香椿', '☀️ 夏季：黄瓜、冬瓜、丝瓜',
                '🍂 秋季：南瓜、排骨、螃蟹', '❄️ 冬季：萝卜、白菜、羊肉'
            ]
        };
    }

    // 智能搜索输入处理
    handleSearchInput(value) {
        this.detectSearchType(value);
        this.updateTypeIndicator();
        this.updateIngredientsDisplay();
        this.updateAutocomplete(value);
    }

    // 检测搜索类型
    detectSearchType(value) {
        // 简单的逻辑：如果包含常见的菜名关键词，则认为是菜品搜索
        const dishKeywords = ['红烧', '宫保', '糖醋', '清蒸', '炒', '煮', '汤', '面', '饭'];
        const isDish = dishKeywords.some(keyword => value.includes(keyword));

        // 如果输入的是逗号分隔的多个词，倾向于认为是食材
        const isIngredients = value.includes(',') || value.includes('、') || value.split(/\s+/).length > 2;

        this.currentSearchType = isDish && !isIngredients ? 'dish' : 'ingredients';
    }

    // 更新类型指示器
    updateTypeIndicator() {
        if (this.currentSearchType === 'ingredients') {
            this.typeBadge.innerHTML = '<i class="bi bi-basket2"></i> 食材模式';
        } else {
            this.typeBadge.innerHTML = '<i class="bi bi-book"></i> 菜品模式';
        }
    }

    // 更新食材显示
    updateIngredientsDisplay() {
        if (this.currentSearchType === 'ingredients') {
            const ingredients = this.extractIngredients(this.smartSearchInput.value);
            if (ingredients.length > 0) {
                this.ingredientsDisplay.style.display = 'block';
                this.displayIngredientsTags(ingredients);
            } else {
                this.ingredientsDisplay.style.display = 'none';
            }
        } else {
            this.ingredientsDisplay.style.display = 'none';
        }
    }

    // 提取食材
    extractIngredients(text) {
        // 支持逗号、顿号、空格分隔
        return text.split(/[,、\s]+/)
            .map(item => item.trim())
            .filter(item => item.length > 0)
            .filter((item, index, arr) => arr.indexOf(item) === index); // 去重
    }

    // 更新食材显示
    updateIngredientsDisplay() {
        if (this.currentSearchType === 'ingredients') {
            const ingredients = this.extractIngredients(this.smartSearchInput.value);
            if (ingredients.length > 0) {
                this.ingredientsDisplay.style.display = 'block';
                this.displayIngredientsTags(ingredients);
            } else {
                this.ingredientsDisplay.style.display = 'none';
            }
        } else {
            this.ingredientsDisplay.style.display = 'none';
        }
    }

    // 显示食材标签
    displayIngredientsTags(ingredients) {
        this.ingredientTagsCompact.innerHTML = '';

        ingredients.forEach((ingredient, originalIndex) => {
            const tag = document.createElement('span');
            tag.className = 'ingredient-tag';
            tag.dataset.ingredient = ingredient;
            tag.dataset.originalIndex = originalIndex;

            // 创建标签内容
            const content = document.createElement('span');
            content.className = 'tag-content';
            content.textContent = ingredient;

            // 创建删除按钮
            const removeBtn = document.createElement('span');
            removeBtn.className = 'remove-tag';
            removeBtn.innerHTML = '<i class="bi bi-x"></i>';
            removeBtn.setAttribute('title', '删除此食材');
            removeBtn.addEventListener('click', (e) => {
                e.stopPropagation();
                this.removeIngredient(ingredient);
            });

            // 设置编辑模式
            tag.addEventListener('dblclick', (e) => {
                e.stopPropagation();
                this.editIngredient(tag, ingredient);
            });

            tag.appendChild(content);
            tag.appendChild(removeBtn);
            this.ingredientTagsCompact.appendChild(tag);

            // 添加进入动画
            setTimeout(() => {
                tag.classList.add('tag-visible');
            }, originalIndex * 50);
        });
    }

    // 优化移除食材标签
    removeIngredient(ingredientName) {
        const currentIngredients = this.extractIngredients(this.smartSearchInput.value);
        const updatedIngredients = currentIngredients.filter(ingredient => ingredient !== ingredientName);

        // 添加删除动画
        const tagToRemove = this.ingredientTagsCompact.querySelector(`[data-ingredient="${ingredientName}"]`);
        if (tagToRemove) {
            tagToRemove.classList.add('tag-removing');
            setTimeout(() => {
                this.smartSearchInput.value = updatedIngredients.join('、');
                this.handleSearchInput(this.smartSearchInput.value);
                this.showRemoveFeedback(ingredientName);
            }, 200);
        } else {
            // 如果没有找到标签元素，直接更新
            this.smartSearchInput.value = updatedIngredients.join('、');
            this.handleSearchInput(this.smartSearchInput.value);
        }
    }

    // 编辑食材标签
    editIngredient(tag, currentName) {
        const content = tag.querySelector('.tag-content');
        const input = document.createElement('input');
        input.type = 'text';
        input.className = 'ingredient-edit-input';
        input.value = currentName;

        const saveEdit = () => {
            const newName = input.value.trim();
            if (newName && newName !== currentName) {
                const currentIngredients = this.extractIngredients(this.smartSearchInput.value);
                const index = currentIngredients.indexOf(currentName);
                if (index !== -1) {
                    currentIngredients[index] = newName;
                    this.smartSearchInput.value = currentIngredients.join('、');
                    this.handleSearchInput(this.smartSearchInput.value);
                    this.showEditFeedback(currentName, newName);
                }
            } else {
                // 恢复原样
                content.style.display = 'inline';
                if (input.parentNode) input.remove();
            }
        };

        input.addEventListener('blur', saveEdit);
        input.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                saveEdit();
            }
        });

        input.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                content.style.display = 'inline';
                if (input.parentNode) input.remove();
            }
        });

        content.style.display = 'none';
        tag.appendChild(input);
        input.focus();
        input.select();
    }

    // 显示删除反馈
    showRemoveFeedback(ingredientName) {
        this.showToast(`已删除食材：${ingredientName}`, 'warning');
    }

    // 显示编辑反馈
    showEditFeedback(oldName, newName) {
        this.showToast(`食材已更新：${oldName} → ${newName}`, 'success');
    }

    // 显示提示信息
    showToast(message, type = 'info') {
        const existingToast = document.querySelector('.toast-message');
        if (existingToast) {
            existingToast.remove();
        }

        const toast = document.createElement('div');
        toast.className = `toast-message toast-${type}`;
        toast.innerHTML = `
            <div class="toast-content">
                <i class="bi bi-${this.getToastIcon(type)}"></i>
                <span>${message}</span>
            </div>
            <button class="toast-close" onclick="this.parentElement.remove()">
                <i class="bi bi-x"></i>
            </button>
        `;

        document.body.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('toast-show');
        }, 10);

        setTimeout(() => {
            toast.classList.remove('toast-show');
            setTimeout(() => {
                if (toast.parentNode) {
                    toast.remove();
                }
            }, 300);
        }, 3000);
    }

    // 获取提示图标
    getToastIcon(type) {
        const icons = {
            'success': 'check-circle-fill',
            'warning': 'exclamation-triangle-fill',
            'error': 'x-circle-fill',
            'info': 'info-circle-fill'
        };
        return icons[type] || icons['info'];
    }

    // 自动补全功能
    updateAutocomplete(value) {
        if (!value.trim()) {
            this.hideAutocomplete();
            return;
        }

        let suggestions = [];
        if (this.currentSearchType === 'ingredients') {
            suggestions = this.getIngredientSuggestions(value);
        } else {
            suggestions = this.getDishSuggestions(value);
        }

        if (suggestions.length > 0) {
            this.showSuggestions(suggestions);
        } else {
            this.hideAutocomplete();
        }
    }

    // 获取食材建议
    getIngredientSuggestions(value) {
        const allIngredients = [
            '鸡蛋', '西红柿', '青菜', '土豆', '萝卜', '猪肉', '牛肉', '鸡肉',
            '鱼肉', '洋葱', '大蒜', '生姜', '葱', '香菜', '豆腐', '豆芽',
            '黄瓜', '冬瓜', '茄子', '青椒', '红椒', '西兰花', '菜花', '白菜'
        ];

        return allIngredients.filter(item =>
            item.toLowerCase().includes(value.toLowerCase()) &&
            !this.extractIngredients(this.smartSearchInput.value).includes(item)
        ).slice(0, 6);
    }

    // 获取菜品建议
    getDishSuggestions(value) {
        const allDishes = [
            '红烧肉', '宫保鸡丁', '糖醋排骨', '麻婆豆腐', '水煮鱼', '清蒸鲈鱼',
            '西红柿鸡蛋汤', '冬瓜排骨汤', '炸酱面', '牛肉面', '刀削面',
            '回锅肉', '鱼香肉丝', '京酱肉丝', '锅包肉', '地三鲜'
        ];

        return allDishes.filter(item =>
            item.toLowerCase().includes(value.toLowerCase())
        ).slice(0, 6);
    }

    // 显示建议
    showSuggestions(suggestions) {
        this.autocompleteItems.innerHTML = '';
        suggestions.forEach(suggestion => {
            const item = document.createElement('div');
            item.className = 'autocomplete-item';
            item.textContent = suggestion;
            item.addEventListener('click', () => {
                this.selectSuggestion(suggestion);
            });
            this.autocompleteItems.appendChild(item);
        });
        this.autocompleteDropdown.style.display = 'block';
    }

    // 选择建议
    selectSuggestion(suggestion) {
        if (this.currentSearchType === 'ingredients') {
            const currentIngredients = this.extractIngredients(this.smartSearchInput.value);
            if (!currentIngredients.includes(suggestion)) {
                currentIngredients.push(suggestion);
                this.smartSearchInput.value = currentIngredients.join('、');
            }
        } else {
            this.smartSearchInput.value = suggestion;
        }

        this.hideAutocomplete();
        this.handleSearchInput(this.smartSearchInput.value);
        this.smartSearchInput.focus();
    }

    // 显示/隐藏自动补全
    showAutocomplete() {
        if (this.smartSearchInput.value.trim()) {
            this.updateAutocomplete(this.smartSearchInput.value);
        }
    }

    hideAutocomplete() {
        this.autocompleteDropdown.style.display = 'none';
    }

    // 处理搜索提交
    handleSearchSubmit() {
        const value = this.smartSearchInput.value.trim();
        if (!value) {
            this.showError('请输入食材或菜品名称');
            return;
        }
        this.performSearch();
    }

    // 智能搜索清空
    clearSearchInput() {
        this.smartSearchInput.value = '';
        this.ingredientsDisplay.style.display = 'none';
        this.hideAutocomplete();
        this.smartSearchInput.focus();
    }

    // 切换提示标签
    switchHintTab(tabElement) {
        // 移除所有活动状态
        this.hintTabs.forEach(tab => tab.classList.remove('active'));

        // 添加活动状态到当前标签
        tabElement.classList.add('active');

        // 更新提示内容
        const hintType = tabElement.dataset.hint;
        this.updateHintContent(hintType);
    }

    // 更新提示内容
    updateHintContent(type) {
        const items = this.hintsData[type] || [];
        const hintHTML = `<div class="hint-items">${items.map(item =>
            `<span class="hint-item">${item}</span>`
        ).join('')}</div>`;
        this.hintContent.innerHTML = hintHTML;

        // 为提示项添加点击事件
        this.hintContent.querySelectorAll('.hint-item').forEach(item => {
            item.addEventListener('click', () => {
                this.selectHint(item.textContent);
            });
        });
    }

    // 选择提示
    selectHint(hintText) {
        // 清理emoji和前缀
        const cleanText = hintText.replace(/[🥚🥬🥩🥕🫛🍄🧀🥦🥘🍛🥣🍜🌸☀️🍂❄️]/g, '').trim();

        if (this.currentSearchType === 'ingredients') {
            const currentIngredients = this.extractIngredients(this.smartSearchInput.value);
            // 分割提示文本中的多个食材
            const hintIngredients = cleanText.split(/[，、:]/).map(item => item.trim()).filter(item => item);
            hintIngredients.forEach(ingredient => {
                if (!currentIngredients.includes(ingredient)) {
                    currentIngredients.push(ingredient);
                }
            });
            this.smartSearchInput.value = currentIngredients.join('、');
        } else {
            this.smartSearchInput.value = cleanText;
        }

        this.handleSearchInput(this.smartSearchInput.value);
    }

    // 显示菜单 (占位符功能)
    showMenu() {
        this.showError('菜单功能正在开发中...');
    }

    // 快速分析 (占位符功能)
    quickAnalysis() {
        if (!this.smartSearchInput.value.trim()) {
            this.showError('请先输入食材或菜品名称');
            return;
        }
        this.performSearch();
    }

    // 开始语音搜索 (占位符功能)
    startVoiceSearch() {
        this.showError('语音搜索功能正在开发中...');
    }

    // 清除结果
    clearResults() {
        this.resultsSection.style.display = 'none';
        this.resultDetails.style.display = 'none';
        this.recipeCardsContainer.innerHTML = '';
        this.smartSearchInput.focus();
    }

    // 隐藏详情
    hideDetails() {
        this.resultDetails.style.display = 'none';
    }

    
    // 执行搜索
    async performSearch() {
        const value = this.smartSearchInput.value.trim();
        if (!value) {
            this.showError('请输入食材或菜品名称');
            return;
        }

        let requestData = {};

        if (this.currentSearchType === 'ingredients') {
            const ingredients = this.extractIngredients(value);
            if (ingredients.length === 0) {
                this.showError('请输入有效的食材');
                return;
            }
            requestData = {
                queryType: 'ingredients',
                ingredients: ingredients
            };
        } else {
            requestData = {
                queryType: 'dish',
                dishName: value
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

        // 随机选择加载图标
        this.selectRandomLoader();

        // 添加步骤指示器动画
        this.animateAnalysisSteps();

        // 添加加载动画文本变化
        this.animateLoadingText();
    }

    // 随机选择加载图标
    selectRandomLoader() {
        const useChef = Math.random() > 0.5;
        if (useChef) {
            this.chefLoader.style.display = 'block';
            this.spoonLoader.style.display = 'none';
        } else {
            this.chefLoader.style.display = 'none';
            this.spoonLoader.style.display = 'block';
        }
    }

    // 动画步骤指示器
    animateAnalysisSteps() {
        this.analysisStepElements.forEach((step, index) => {
            step.classList.remove('active', 'completed');
        });

        let currentStep = 0;

        const stepInterval = setInterval(() => {
            // 完成当前步骤
            if (currentStep > 0) {
                this.analysisStepElements[currentStep - 1].classList.remove('active');
                this.analysisStepElements[currentStep - 1].classList.add('completed');
            }

            // 激活当前步骤
            if (currentStep < this.analysisStepElements.length) {
                this.analysisStepElements[currentStep].classList.add('active');
                currentStep++;
            } else {
                clearInterval(stepInterval);
            }
        }, 800);

        this.stepInterval = stepInterval;
    }

    // 动画加载文本
    animateLoadingText() {
        const messages = [
            '👨‍🍳 AI厨师正在分析食材...',
            '🔥 激活智能烹饪算法...',
            '🥘 匹配最佳食谱方案...',
            '🥄 计算营养和烹饪要点...',
            '✨ 生成个性化建议...'
        ];
        let index = 0;

        // 删除旧的间隔
        if (this.loadingInterval) {
            clearInterval(this.loadingInterval);
        }

        // 打字效果
        const typeMessage = (message, callback) => {
            let charIndex = 0;
            this.dynamicLoadingText.textContent = '';

            const typeInterval = setInterval(() => {
                if (charIndex < message.length) {
                    this.dynamicLoadingText.textContent += message[charIndex];
                    charIndex++;
                } else {
                    clearInterval(typeInterval);
                    if (callback) callback();
                }
            }, 30);
        };

        const showNextMessage = () => {
            typeMessage(messages[index], () => {
                index = (index + 1) % messages.length;
                if (this.loadingContainer.style.display === 'block') {
                    setTimeout(showNextMessage, 1500);
                }
            });
        };

        showNextMessage();
    }

    // 隐藏加载状态
    hideLoading() {
        if (this.loadingInterval) {
            clearInterval(this.loadingInterval);
        }
        if (this.stepInterval) {
            clearInterval(this.stepInterval);
        }

        this.loadingContainer.style.display = 'none';
        this.searchBtn.disabled = false;
        this.clearBtn.disabled = false;
    }

    // 显示结果
    showResult(data) {
        this.hideAllContainers();
        this.resultsSection.style.display = 'block';

        // 设置元数据
        this.serveCount.textContent = this.getQueryTypeLabel(data.type);
        this.timestamp.textContent = this.formatTimestamp(data.timestamp);

        // 创建卡片式结果
        this.createRecipeCards(data);

        // 显示附加信息
        if (data.supplementaryData && Object.keys(data.supplementaryData).length > 0) {
            this.showSupplementaryData(data.supplementaryData);
        }

        // 先移除动画类，再添加以确保动画重新触发
        this.resultsSection.classList.remove('show');

        // 触发重绘以重新开始动画
        void this.resultsSection.offsetWidth;

        // 添加动画效果
        this.resultsSection.classList.add('show');

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
        this.resultsSection.style.display = 'none';
        this.errorContainer.style.display = 'none';
        this.resultDetails.style.display = 'none';
    }

    // 清空所有内容
    clearAll() {
        this.smartSearchInput.value = '';
        this.hideAllContainers();
        this.smartSearchInput.focus();
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

    // 创建菜谱卡片
    createRecipeCards(data) {
        this.recipeCardsContainer.innerHTML = '';

        // 如果是食材分析，创建一个汇总卡片
        if (data.type === 'ingredients') {
            const summaryCard = this.createSummaryCard(data);
            this.recipeCardsContainer.appendChild(summaryCard);
        }

        // 创建详细内容卡片
        const detailCard = this.createDetailCard(data);
        this.recipeCardsContainer.appendChild(detailCard);

        // 如果有API食谱，创建食谱卡片数组
        if (data.supplementaryData && data.supplementaryData.api_recipes) {
            data.supplementaryData.api_recipes.forEach((recipe, index) => {
                const recipeCard = this.createApiRecipeCard(recipe);
                this.recipeCardsContainer.appendChild(recipeCard);
            });
        }
    }

    // 创建汇总卡片
    createSummaryCard(data) {
        const card = document.createElement('div');
        card.className = 'recipe-card';

        const ingredients = this.currentSearchType === 'ingredients'
            ? this.extractIngredients(this.smartSearchInput.value).join('、')
            : this.smartSearchInput.value;

        card.innerHTML = `
            <div class="recipe-card-image">
                <img src="https://via.placeholder.com/320x200/66BB6A/ffffff?text=食材分析" alt="食材分析">
                <div class="recipe-card-badge">AI 分析</div>
            </div>
            <div class="recipe-card-content">
                <h3 class="recipe-card-title">智能分析结果</h3>
                <div class="recipe-card-meta">
                    <div class="recipe-card-meta-item">
                        <i class="bi bi-basket2"></i>
                        ${ingredients}
                    </div>
                </div>
                <div class="recipe-card-rating">
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-half star-rating"></i>
                    <span>4.5</span>
                </div>
                <p class="recipe-card-description">
                    AI 为您分析了 ${this.currentSearchType === 'ingredients' ? '食材组合' : '菜品特色'}，
                    提供详细的制作指导和营养建议。
                </p>
            </div>
        `;

        card.addEventListener('click', () => {
            this.showDetailedResult(data);
        });

        return card;
    }

    // 创建详细内容卡片
    createDetailCard(data) {
        const card = document.createElement('div');
        card.className = 'recipe-card';

        const preview = this.getPreviewText(data.result);

        card.innerHTML = `
            <div class="recipe-card-image">
                <img src="https://via.placeholder.com/320x200/FF6B6B/ffffff?text=详细内容" alt="详细内容">
                <div class="recipe-card-badge">详细步骤</div>
            </div>
            <div class="recipe-card-content">
                <h3 class="recipe-card-title">制作指南</h3>
                <div class="recipe-card-meta">
                    <div class="recipe-card-meta-item">
                        <i class="bi bi-clock"></i>
                        约30分钟
                    </div>
                    <div class="recipe-card-meta-item">
                        <i class="bi bi-people"></i>
                        2-3人份
                    </div>
                </div>
                <p class="recipe-card-description">${preview}</p>
            </div>
        `;

        card.addEventListener('click', () => {
            this.showDetailedResult(data);
        });

        return card;
    }

    // 创建API食谱卡片
    createApiRecipeCard(recipe) {
        const card = document.createElement('div');
        card.className = 'recipe-card';

        card.innerHTML = `
            <div class="recipe-card-image">
                <img src="https://via.placeholder.com/320x200/FFA726/ffffff?text=${encodeURIComponent(recipe.title)}" alt="${recipe.title}">
                ${recipe.readyInMinutes ? `<div class="recipe-card-badge">${recipe.readyInMinutes}分钟</div>` : ''}
            </div>
            <div class="recipe-card-content">
                <h3 class="recipe-card-title">${recipe.title}</h3>
                <div class="recipe-card-meta">
                    ${recipe.readyInMinutes ? `
                        <div class="recipe-card-meta-item">
                            <i class="bi bi-clock"></i>
                            ${recipe.readyInMinutes}分钟
                        </div>
                    ` : ''}
                    ${recipe.servings ? `
                        <div class="recipe-card-meta-item">
                            <i class="bi bi-people"></i>
                            ${recipe.servings}人份
                        </div>
                    ` : ''}
                </div>
                <div class="recipe-card-rating">
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star-fill star-rating"></i>
                    <i class="bi bi-star star-rating"></i>
                    <span>4.0</span>
                </div>
                <p class="recipe-card-description">
                    来自专业食谱数据库的经典做法，包含详细的步骤和营养信息。
                </p>
            </div>
        `;

        card.addEventListener('click', () => {
            this.showRecipeDetails(recipe);
        });

        return card;
    }

    // 获取预览文本
    getPreviewText(content) {
        if (!content) return '暂无内容';

        // 移除HTML标签并截取前150个字符
        const textContent = content
            .replace(/<[^>]*>/g, '')
            .replace(/\n+/g, ' ')
            .trim();

        return textContent.length > 150
            ? textContent.substring(0, 150) + '...'
            : textContent;
    }

    // 显示详细结果
    showDetailedResult(data) {
        this.resultDetails.style.display = 'block';
        this.detailsContent.innerHTML = this.formatContent(data.result);

        // 滚动到详情区域
        this.resultDetails.scrollIntoView({
            behavior: 'smooth',
            block: 'start'
        });
    }

    // 显示食谱详情
    showRecipeDetails(recipe) {
        this.resultDetails.style.display = 'block';

        let detailsHTML = `
            <h3>${recipe.title}</h3>
            <div class="recipe-details-meta">
                ${recipe.readyInMinutes ? `<p><i class="bi bi-clock"></i> 制作时间：${recipe.readyInMinutes}分钟</p>` : ''}
                ${recipe.servings ? `<p><i class="bi bi-people"></i> 分量：${recipe.servings}人份</p>` : ''}
            </div>
            <p>详细的制作步骤和营养信息正在加载中...</p>
        `;

        this.detailsContent.innerHTML = detailsHTML;
        this.resultDetails.scrollIntoView({
            behavior: 'smooth',
            block: 'start'
        });
    }

    // 滚动到结果区域
    scrollToResult() {
        setTimeout(() => {
            this.resultsSection.scrollIntoView({
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