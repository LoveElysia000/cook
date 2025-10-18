package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

// RecipeService 食谱服务结构
type RecipeService struct {
	apiKey     string
	baseURL    string
	cache      map[string]*CacheEntry
	cacheMutex sync.RWMutex
}

// CacheEntry 缓存条目
type CacheEntry struct {
	Data      string
	ExpiresAt time.Time
}

// SpoonacularRecipe Spoonacular API食谱结构
type SpoonacularRecipe struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Image       string            `json:"image"`
	ImageType   string            `json:"imageType"`
	Instructions string            `json:"instructions"`
	Servings    int              `json:"servings"`
	ReadyInMinutes int            `json:"readyInMinutes"`
	ExtendedIngredients []ExtendedIngredient `json:"extendedIngredients"`
}

// ExtendedIngredient 扩展食材
type ExtendedIngredient struct {
	ID     int     `json:"id"`
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

// SpoonacularResponse API响应结构
type SpoonacularResponse struct {
	Results []SpoonacularRecipe `json:"results"`
	Offset  int                 `json:"offset"`
	Number  int                 `json:"number"`
	Total   int                 `json:"totalResults"`
}

// NewRecipeService 创建食谱服务实例
func NewRecipeService() *RecipeService {
	return &RecipeService{
		apiKey:  os.Getenv("SPOONACULAR_API_KEY"),
		baseURL: "https://api.spoonacular.com/recipes",
		cache:   make(map[string]*CacheEntry),
	}
}

// IngredientTranslation 食材中英文翻译映射
var IngredientTranslation = map[string]string{
	// 蛋类
	"鸡蛋": "eggs",
	"蛋": "eggs",
	"鸭蛋": "duck eggs",
	"鹌鹑蛋": "quail eggs",

	// 蔬菜类
	"西红柿": "tomato",
	"番茄": "tomato",
	"土豆": "potato",
	"马铃薯": "potato",
	"黄瓜": "cucumber",
	"青椒": "bell pepper",
	"红椒": "red bell pepper",
	"胡萝卜": "carrot",
	"白菜": "cabbage",
	"菠菜": "spinach",
	"芹菜": "celery",
	"洋葱": "onion",
	"蒜": "garlic",
	"生姜": "ginger",
	"姜": "ginger",
	"韭菜": "chives",
	"豆芽": "bean sprouts",
	"蘑菇": "mushroom",
	"香菇": "shiitake mushroom",
	"金针菇": "enoki mushroom",

	// 肉类
	"猪肉": "pork",
	"牛肉": "beef",
	"鸡肉": "chicken",
	"鸭肉": "duck",
	"羊肉": "lamb",
	"鱼肉": "fish",
	"虾": "shrimp",
	"蟹": "crab",

	// 调料类
	"葱": "green onion",
	"小葱": "green onion",
	"香菜": "cilantro",
	"八角": "star anise",
	"桂皮": "cinnamon",
	"花椒": "sichuan peppercorn",
	"料酒": "cooking wine",
	"生抽": "light soy sauce",
	"老抽": "dark soy sauce",
	"陈醋": "balsamic vinegar",
	"香油": "sesame oil",
	"盐": "salt",
	"糖": "sugar",
	"白糖": "white sugar",
	"冰糖": "rock sugar",

	// 其他
	"大米": "rice",
	"面条": "noodles",
	"豆腐": "tofu",
	"豆浆": "soy milk",
	"牛奶": "milk",
	"面包": "bread",
	"奶酪": "cheese",
}

// translateIngredients 翻译中文食材为英文
func (s *RecipeService) translateIngredients(ingredients []string) []string {
	var translated []string
	for _, ingredient := range ingredients {
		// 如果是中文，则翻译成英文；如果是英文，直接使用
		if english, exists := IngredientTranslation[ingredient]; exists {
			translated = append(translated, english)
		} else {
			// 检查是否包含中文字符，如果包含但没有翻译映射，则跳过
			if s.containsChinese(ingredient) {
				// 可以跳过或使用原词，这里选择跳过避免API异常
				continue
			} else {
				// 英文食材直接使用
				translated = append(translated, ingredient)
			}
		}
	}
	return translated
}

// containsChinese 检查字符串是否包含中文字符
func (s *RecipeService) containsChinese(str string) bool {
	for _, r := range str {
		if r >= rune('\u4e00') && r <= rune('\u9fff') {
			return true
		}
	}
	return false
}

// SearchByIngredients 根据食材搜索食谱
func (s *RecipeService) SearchByIngredients(ingredients []string) ([]SpoonacularRecipe, error) {
	if s.apiKey == "" {
		return []SpoonacularRecipe{}, nil
	}

	// 翻译中文食材为英文
	translatedIngredients := s.translateIngredients(ingredients)
	if len(translatedIngredients) == 0 {
		return []SpoonacularRecipe{}, nil
	}

	// 检查缓存（使用原始食材作为缓存键）
	cacheKey := s.generateCacheKey("ingredients", ingredients)
	if cached := s.getFromCache(cacheKey); cached != "" {
		var recipes []SpoonacularRecipe
		if err := json.Unmarshal([]byte(cached), &recipes); err == nil {
			return recipes, nil
		}
	}

	// 构建请求参数（使用翻译后的英文食材）
	ingredientsStr := strings.Join(translatedIngredients, ",+")
	apiURL := fmt.Sprintf("%s/findByIngredients?ingredients=%s&number=5&apiKey=%s",
		s.baseURL, url.QueryEscape(ingredientsStr), s.apiKey)

	// 发送请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return []SpoonacularRecipe{}, fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return []SpoonacularRecipe{}, fmt.Errorf("API返回错误: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []SpoonacularRecipe{}, fmt.Errorf("读取响应失败: %v", err)
	}

	var recipes []SpoonacularRecipe
	if err := json.Unmarshal(body, &recipes); err != nil {
		return []SpoonacularRecipe{}, fmt.Errorf("解析响应失败: %v", err)
	}

	// 缓存结果
	s.saveToCache(cacheKey, string(body), 30*time.Minute)

	return recipes, nil
}

// DishTranslation 菜品中英文翻译映射
var DishTranslation = map[string]string{
	// 基础菜品
	"西红柿炒鸡蛋": "scrambled eggs with tomato",
	"番茄炒蛋": "scrambled eggs with tomato",
	"炒鸡蛋": "scrambled eggs",
	"蒸蛋": "steamed egg",
	"煎蛋": "fried egg",
	"蛋汤": "egg soup",
	"土豆丝": "shredded potato",
	"炒土豆丝": "stir-fried shredded potato",
	"红烧肉": "braised pork belly",
	"糖醋排骨": "sweet and sour pork ribs",
	"宫保鸡丁": "kung pao chicken",
	"麻婆豆腐": "mapo tofu",
	"鱼香肉丝": "fish-flavored pork",
	"青椒肉丝": "stir-fried pork with green pepper",
	"水煮鱼": "boiled fish",
	"回锅肉": "twice-cooked pork",
	"酸菜鱼": "fish with pickled vegetables",
	"火锅": "hot pot",
	"饺子": "dumplings",
	"面条": "noodles",
	"炒面": "fried noodles",
	"汤面": "noodle soup",
	"炒饭": "fried rice",
	"蛋炒饭": "egg fried rice",
	"粥": "congee",
	"白粥": "plain congee",
	"皮蛋瘦肉粥": "preserved egg and pork congee",
}

// translateDishName 翻译中文菜名为英文
func (s *RecipeService) translateDishName(dishName string) string {
	// 如果是中文菜名，则翻译成英文；如果是英文菜名，直接使用
	if english, exists := DishTranslation[dishName]; exists {
		return english
	} else {
		// 检查是否包含中文字符
		if s.containsChinese(dishName) {
			// 对于没有翻译映射的中文菜名，可以通过关键词匹配
			if strings.Contains(dishName, "鸡蛋") || strings.Contains(dishName, "炒蛋") {
				return "scrambled eggs"
			} else if strings.Contains(dishName, "土豆") {
				return "potato"
			} else if strings.Contains(dishName, "鸡肉") || strings.Contains(dishName, "鸡") {
				return "chicken"
			} else if strings.Contains(dishName, "猪肉") || strings.Contains(dishName, "肉") {
				return "pork"
			} else if strings.Contains(dishName, "鱼") {
				return "fish"
			} else if strings.Contains(dishName, "豆腐") {
				return "tofu"
			} else if strings.Contains(dishName, "面条") || strings.Contains(dishName, "面") {
				return "noodles"
			} else if strings.Contains(dishName, "米饭") || strings.Contains(dishName, "饭") {
				return "rice"
			} else {
				// 包含中文但无法识别的，返回空字符串避免API异常
				return ""
			}
		} else {
			// 英文菜名直接使用
			return dishName
		}
	}
}

// SearchByDishName 根据菜品名搜索食谱
func (s *RecipeService) SearchByDishName(dishName string) ([]SpoonacularRecipe, error) {
	if s.apiKey == "" {
		return []SpoonacularRecipe{}, nil
	}

	// 翻译中文菜名为英文
	translatedDishName := s.translateDishName(dishName)
	if translatedDishName == "" {
		return []SpoonacularRecipe{}, nil
	}

	// 检查缓存（使用原始菜名作为缓存键）
	cacheKey := s.generateCacheKey("dish", []string{dishName})
	if cached := s.getFromCache(cacheKey); cached != "" {
		var recipes []SpoonacularRecipe
		if err := json.Unmarshal([]byte(cached), &recipes); err == nil {
			return recipes, nil
		}
	}

	// 构建请求参数（使用翻译后的英文菜名）
	apiURL := fmt.Sprintf("%s/complexSearch?query=%s&number=5&addRecipeInformation=true&apiKey=%s",
		s.baseURL, url.QueryEscape(translatedDishName), s.apiKey)

	// 发送请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return []SpoonacularRecipe{}, fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return []SpoonacularRecipe{}, fmt.Errorf("API返回错误: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []SpoonacularRecipe{}, fmt.Errorf("读取响应失败: %v", err)
	}

	var searchResp SpoonacularResponse
	if err := json.Unmarshal(body, &searchResp); err != nil {
		return []SpoonacularRecipe{}, fmt.Errorf("解析响应失败: %v", err)
	}

	// 缓存结果
	s.saveToCache(cacheKey, string(body), 60*time.Minute)

	return searchResp.Results, nil
}

// GetRecipeInformation 获取详细食谱信息
func (s *RecipeService) GetRecipeInformation(recipeID int) (*SpoonacularRecipe, error) {
	if s.apiKey == "" {
		return nil, fmt.Errorf("API密钥未配置")
	}

	// 检查缓存
	cacheKey := fmt.Sprintf("recipe_info_%d", recipeID)
	if cached := s.getFromCache(cacheKey); cached != "" {
		var recipe SpoonacularRecipe
		if err := json.Unmarshal([]byte(cached), &recipe); err == nil {
			return &recipe, nil
		}
	}

	// 构建请求参数
	apiURL := fmt.Sprintf("%s/%d/information?includeNutrition=false&apiKey=%s",
		s.baseURL, recipeID, s.apiKey)

	// 发送请求
	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("API请求失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API返回错误: %d - %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var recipe SpoonacularRecipe
	if err := json.Unmarshal(body, &recipe); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 缓存结果
	s.saveToCache(cacheKey, string(body), 60*time.Minute)

	return &recipe, nil
}

// FormatRecipesForAI 将食谱格式化为AI可读取的格式
func (s *RecipeService) FormatRecipesForAI(recipes []SpoonacularRecipe) string {
	if len(recipes) == 0 {
		return "未找到相关食谱参考。"
	}

	var result strings.Builder
	result.WriteString("## 参考食谱信息\n\n")

	for i, recipe := range recipes {
		result.WriteString(fmt.Sprintf("### 参考食谱 %d: %s\n", i+1, recipe.Title))

		if recipe.ReadyInMinutes > 0 {
			result.WriteString(fmt.Sprintf("- **制作时间**: %d分钟\n", recipe.ReadyInMinutes))
		}

		if recipe.Servings > 0 {
			result.WriteString(fmt.Sprintf("- **份量**: %d人份\n", recipe.Servings))
		}

		if len(recipe.ExtendedIngredients) > 0 {
			result.WriteString("- **食材**:\n")
			for _, ing := range recipe.ExtendedIngredients {
				result.WriteString(fmt.Sprintf("  - %s: %.1f %s\n", ing.Name, ing.Amount, ing.Unit))
			}
		}

		if recipe.Instructions != "" {
			result.WriteString("- **制作说明**: <制作步骤参考>\n")
		}

		result.WriteString("\n")
	}

	return result.String()
}

// generateCacheKey 生成缓存键
func (s *RecipeService) generateCacheKey(prefix string, items []string) string {
	return fmt.Sprintf("%s:%s", prefix, strings.Join(items, "|"))
}

// getFromCache 从缓存获取数据
func (s *RecipeService) getFromCache(key string) string {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	if entry, exists := s.cache[key]; exists {
		if time.Now().Before(entry.ExpiresAt) {
			return entry.Data
		}
		// 清理过期缓存
		delete(s.cache, key)
	}
	return ""
}

// saveToCache 保存数据到缓存
func (s *RecipeService) saveToCache(key, data string, ttl time.Duration) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	s.cache[key] = &CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// CleanExpiredCache 清理过期缓存
func (s *RecipeService) CleanExpiredCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	now := time.Now()
	for key, entry := range s.cache {
		if now.After(entry.ExpiresAt) {
		delete(s.cache, key)
		}
	}
}

// GetCacheStatus 获取缓存状态
func (s *RecipeService) GetCacheStatus() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()

	now := time.Now()
	activeCount := 0
	expiredCount := 0

	for _, entry := range s.cache {
		if now.Before(entry.ExpiresAt) {
			activeCount++
		} else {
			expiredCount++
		}
	}

	return map[string]interface{}{
		"total_entries":  len(s.cache),
		"active_entries": activeCount,
		"expired_entries": expiredCount,
	}
}