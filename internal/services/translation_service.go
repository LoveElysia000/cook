package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// TranslationService 翻译服务结构
type TranslationService struct {
	aiAPIKey     string
	aiBaseURL    string
	model        string
	cache        map[string]*TranslationCacheEntry
	cacheMutex   sync.RWMutex
	// 保留高频常用词的静态映射作为快速查询
	commonTranslations map[string]string
}

// TranslationCacheEntry 翻译缓存条目
type TranslationCacheEntry struct {
	Translation string
	ExpiresAt   time.Time
}

// NewTranslationService 创建翻译服务实例
func NewTranslationService() *TranslationService {
	return &TranslationService{
		aiAPIKey:  os.Getenv("DEEPSEEK_API_KEY"),
		aiBaseURL: "https://api.deepseek.com/v1/chat/completions",
		model:     "deepseek-chat",
		cache:     make(map[string]*TranslationCacheEntry),
		commonTranslations: map[string]string{
			// 保留最常用的几个快速映射
			"鸡蛋": "eggs",
			"西红柿": "tomato",
			"土豆": "potato",
			"猪肉": "pork",
			"牛肉": "beef",
			"鸡肉": "chicken",
			"鱼肉": "fish",
			"米饭": "rice",
			"面条": "noodles",
			"豆腐": "tofu",
		},
	}
}

// TranslateIngredient 翻译食材名称（中译英）
func (t *TranslationService) TranslateIngredient(ingredient string) string {
	// 1. 检查是否为英文，直接返回
	if !t.containsChinese(ingredient) {
		return ingredient
	}

	// 2. 检查常用翻译映射（快速查询）
	if translation, exists := t.commonTranslations[ingredient]; exists {
		return translation
	}

	// 3. 检查缓存
	cacheKey := "ingredient:" + ingredient
	if cached := t.getFromCache(cacheKey); cached != "" {
		return cached
	}

	// 4. 调用AI进行翻译
	translation, err := t.translateWithAI(ingredient, "ingredient")
	if err != nil {
		// AI翻译失败，尝试关键词匹配作为降级策略
		return t.fallbackTranslation(ingredient)
	}

	// 5. 缓存翻译结果
	t.saveToCache(cacheKey, translation, 24*time.Hour) // 缓存24小时

	return translation
}

// TranslateDishName 翻译菜名（中译英）
func (t *TranslationService) TranslateDishName(dishName string) string {
	// 1. 检查是否为英文，直接返回
	if !t.containsChinese(dishName) {
		return dishName
	}

	// 2. 检查常用翻译映射（快速查询）
	if translation, exists := t.commonTranslations[dishName]; exists {
		return translation
	}

	// 3. 检查缓存
	cacheKey := "dish:" + dishName
	if cached := t.getFromCache(cacheKey); cached != "" {
		return cached
	}

	// 4. 调用AI进行翻译
	translation, err := t.translateWithAI(dishName, "dish")
	if err != nil {
		// AI翻译失败，尝试关键词匹配作为降级策略
		return t.fallbackTranslation(dishName)
	}

	// 5. 缓存翻译结果
	t.saveToCache(cacheKey, translation, 24*time.Hour) // 缓存24小时

	return translation
}

// TranslateIngredients 批量翻译食材
func (t *TranslationService) TranslateIngredients(ingredients []string) []string {
	var translated []string
	for _, ingredient := range ingredients {
		translation := t.TranslateIngredient(ingredient)
		if translation != "" {
			translated = append(translated, translation)
		}
	}
	return translated
}

// translateWithAI 使用AI进行翻译
func (t *TranslationService) translateWithAI(text string, textType string) (string, error) {
	if t.aiAPIKey == "" {
		return "", fmt.Errorf("AI API密钥未配置")
	}

	var prompt string
	if textType == "ingredient" {
		prompt = fmt.Sprintf(`请将以下中文食材名称翻译成英文，只需返回单个英文单词或词组，不要任何解释：
%s`, text)
	} else {
		prompt = fmt.Sprintf(`请将以下中文菜名翻译成对应的英文菜名，返回适合食谱搜索的英文表达：
%s`, text)
	}

	// 构建AI请求
	requestBody := map[string]interface{}{
		"model": t.model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": prompt,
			},
		},
		"stream": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化翻译请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", t.aiBaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建翻译请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.aiAPIKey)

	client := &http.Client{
		Timeout: 10 * time.Second, // 设置10秒超时
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("翻译API调用失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("翻译API返回错误 %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取翻译响应失败: %v", err)
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("解析翻译响应失败: %v", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("翻译API返回空响应")
	}

	// 清理返回的翻译结果
	translation := strings.TrimSpace(apiResp.Choices[0].Message.Content)
	translation = strings.Trim(translation, `"'`) // 去除可能的引号
	translation = strings.ToLower(translation)   // 统一转为小写

	return translation, nil
}

// fallbackTranslation 降级翻译策略
func (t *TranslationService) fallbackTranslation(text string) string {
	// 关键词匹配作为最后的备选方案
	if strings.Contains(text, "鸡蛋") || strings.Contains(text, "炒蛋") {
		return "scrambled eggs"
	} else if strings.Contains(text, "土豆") {
		return "potato"
	} else if strings.Contains(text, "鸡肉") || strings.Contains(text, "鸡") {
		return "chicken"
	} else if strings.Contains(text, "猪肉") || strings.Contains(text, "肉") {
		return "pork"
	} else if strings.Contains(text, "牛肉") {
		return "beef"
	} else if strings.Contains(text, "鱼") {
		return "fish"
	} else if strings.Contains(text, "豆腐") {
		return "tofu"
	} else if strings.Contains(text, "面条") || strings.Contains(text, "面") {
		return "noodles"
	} else if strings.Contains(text, "米饭") || strings.Contains(text, "饭") {
		return "rice"
	} else if strings.Contains(text, "西红柿") || strings.Contains(text, "番茄") {
		return "tomato"
	} else if strings.Contains(text, "青菜") || strings.Contains(text, "菜") {
		return "vegetables"
	} else {
		return "" // 无法翻译则返回空字符串
	}
}

// containsChinese 检查字符串是否包含中文字符
func (t *TranslationService) containsChinese(str string) bool {
	for _, r := range str {
		if r >= rune('\u4e00') && r <= rune('\u9fff') {
			return true
		}
	}
	return false
}

// getFromCache 从缓存获取翻译结果
func (t *TranslationService) getFromCache(key string) string {
	t.cacheMutex.RLock()
	defer t.cacheMutex.RUnlock()

	if entry, exists := t.cache[key]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			return entry.Translation
		}
		// 清理过期缓存
		delete(t.cache, key)
	}
	return ""
}

// saveToCache 保存翻译结果到缓存
func (t *TranslationService) saveToCache(key, translation string, ttl time.Duration) {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()

	t.cache[key] = &TranslationCacheEntry{
		Translation: translation,
		ExpiresAt:   time.Now().Add(ttl),
	}
}

// CleanExpiredCache 清理过期的翻译缓存
func (t *TranslationService) CleanExpiredCache() {
	t.cacheMutex.Lock()
	defer t.cacheMutex.Unlock()

	now := time.Now()
	for key, entry := range t.cache {
		if now.After(entry.ExpiresAt) {
			delete(t.cache, key)
		}
	}
}

// GetCacheStatus 获取翻译缓存状态
func (t *TranslationService) GetCacheStatus() map[string]interface{} {
	t.cacheMutex.RLock()
	defer t.cacheMutex.RUnlock()

	now := time.Now()
	activeCount := 0
	expiredCount := 0

	for _, entry := range t.cache {
		if now.Before(entry.ExpiresAt) {
			activeCount++
		} else {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"total_entries":  len(t.cache),
		"active_entries": activeCount,
		"expired_entries": expiredCount,
	}
}