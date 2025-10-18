package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"recipe-agent/internal/services"
)

// AgentHandler 处理器结构
type AgentHandler struct {
	recipeService *services.RecipeService
	aiService     *services.AIService
}

// RecipeRequest 食谱请求结构
type RecipeRequest struct {
	Ingredients []string `json:"ingredients"`
	DishName    string   `json:"dishName"`
	QueryType   string   `json:"queryType"`
}

// RecipeResponse 食谱响应结构
type RecipeResponse struct {
	Result           string                 `json:"result"`
	Type             string                 `json:"type"`
	Timestamp        time.Time             `json:"timestamp"`
	SupplementaryData map[string]interface{} `json:"supplementaryData"`
	Success          bool                  `json:"success"`
	Message          string                `json:"message,omitempty"`
}

// NewAgentHandler 创建处理器实例
func NewAgentHandler(recipeService *services.RecipeService, aiService *services.AIService) *AgentHandler {
	return &AgentHandler{
		recipeService: recipeService,
		aiService:     aiService,
	}
}

// GetRecipes 处理食谱请求
func (h *AgentHandler) GetRecipes(c *gin.Context) {
	var req RecipeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, RecipeResponse{
			Success: false,
			Message: "请求格式错误: " + err.Error(),
		})
		return
	}

	// 验证请求
	if err := h.validateRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, RecipeResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// 处理请求
	result, supplementaryData, err := h.processRequest(&req)
	if err != nil {
		log.Printf("处理请求失败: %v", err)
		c.JSON(http.StatusInternalServerError, RecipeResponse{
			Success: false,
			Message: "服务处理失败，请稍后重试",
		})
		return
	}

	// 返回结果
	response := RecipeResponse{
		Result:           result,
		Type:             req.QueryType,
		Timestamp:        time.Now(),
		SupplementaryData: supplementaryData,
		Success:          true,
	}

	c.JSON(http.StatusOK, response)
}

// validateRequest 验证请求
func (h *AgentHandler) validateRequest(req *RecipeRequest) error {
	// 验证queryType
	if req.QueryType != "ingredients" && req.QueryType != "dish" {
		return fmt.Errorf("无效的查询类型，必须是 'ingredients' 或 'dish'")
	}

	// 根据查询类型验证相应字段
	switch req.QueryType {
	case "ingredients":
		if len(req.Ingredients) == 0 {
			return fmt.Errorf("按食材查询时必须提供食材列表")
		}
		// 清理和验证食材
		var cleanedIngredients []string
		for _, ingredient := range req.Ingredients {
			ingredient = strings.TrimSpace(ingredient)
			if ingredient != "" {
				cleanedIngredients = append(cleanedIngredients, ingredient)
			}
		}
		if len(cleanedIngredients) == 0 {
			return fmt.Errorf("食材列表不能为空")
		}
		req.Ingredients = cleanedIngredients

	case "dish":
		if strings.TrimSpace(req.DishName) == "" {
			return fmt.Errorf("按菜名查询时必须提供菜名")
		}
		req.DishName = strings.TrimSpace(req.DishName)
	}

	return nil
}

// processRequest 处理具体的食谱请求
func (h *AgentHandler) processRequest(req *RecipeRequest) (string, map[string]interface{}, error) {
	var result string
	var supplementaryData map[string]interface{}
	var err error

	// 处理不同类型的请求
	switch req.QueryType {
	case "ingredients":
		result, supplementaryData, err = h.processIngredientsRequest(req.Ingredients)
	case "dish":
		result, supplementaryData, err = h.processDishRequest(req.DishName)
	default:
		return "", nil, fmt.Errorf("不支持的查询类型: %s", req.QueryType)
	}

	if err != nil {
		return "", nil, err
	}

	return result, supplementaryData, nil
}

// processIngredientsRequest 处理食材请求
func (h *AgentHandler) processIngredientsRequest(ingredients []string) (string, map[string]interface{}, error) {
	log.Printf("处理食材查询请求: %v", ingredients)

	// 并行获取AI分析和API数据
	type apiResult struct {
		recipes []services.SpoonacularRecipe
		err     error
	}

	aiChan := make(chan string, 1)
	apiChan := make(chan apiResult, 1)

	// 异步调用AI服务
	go func() {
		aiResult, err := h.aiService.AnalyzeIngredients(ingredients)
		if err != nil {
			log.Printf("AI服务调用失败: %v", err)
			aiChan <- "" // 发送空结果表示失败
		} else {
			aiChan <- aiResult
		}
	}()

	// 异步调用API服务
	go func() {
		recipes, err := h.recipeService.SearchByIngredients(ingredients)
		apiChan <- apiResult{recipes: recipes, err: err}
	}()

	// 等待AI结果
	aiResult := <-aiChan
	var aiError error
	if aiResult == "" {
		aiError = fmt.Errorf("AI服务不可用")
	}

	// 等待API结果
	apiRes := <-apiChan
	var apiError error
	if apiRes.err != nil {
		apiError = fmt.Errorf("API服务不可用")
	}

	// 构建最终结果
	var finalResult string
	var supplementaryData map[string]interface{}

	if aiError != nil && apiError != nil {
		// 两个服务都失败
		return "", nil, fmt.Errorf("所有服务都不可用")
	} else if aiError == nil && apiError != nil {
		// 只有AI服务可用
		finalResult = aiResult
		supplementaryData = map[string]interface{}{
			"ai_available":    true,
			"api_available":   false,
			"nutrition_tips":  "营养分析暂不可用",
		}
	} else if aiError != nil && apiError == nil {
		// 只有API服务可用
		recipesText := h.recipeService.FormatRecipesForAI(apiRes.recipes)
		finalResult = h.generateFallbackIngredientResult(ingredients, recipesText)
		supplementaryData = map[string]interface{}{
			"ai_available":    false,
			"api_available":   true,
			"api_recipes":     apiRes.recipes,
			"nutrition_tips":  "获得" + fmt.Sprintf("%d", len(apiRes.recipes)) + "个食谱参考",
		}
	} else {
		// 两个服务都可用，整合结果
		finalResult = h.combineIngredientResults(aiResult, apiRes.recipes)
		supplementaryData = map[string]interface{}{
			"ai_available":    true,
			"api_available":   true,
			"api_recipes":     apiRes.recipes,
			"nutrition_tips":  "AI分析完成，包含" + fmt.Sprintf("%d", len(apiRes.recipes)) + "个食谱参考",
		}
	}

	return finalResult, supplementaryData, nil
}

// processDishRequest 处理菜品请求
func (h *AgentHandler) processDishRequest(dishName string) (string, map[string]interface{}, error) {
	log.Printf("处理菜品查询请求: %s", dishName)

	// 并行获取AI详细分析和API数据
	type apiResult struct {
		recipes []services.SpoonacularRecipe
		err     error
	}

	aiChan := make(chan string, 1)
	apiChan := make(chan apiResult, 1)

	// 异步调用AI服务
	go func() {
		aiResult, err := h.aiService.GetDishDetails(dishName)
		if err != nil {
			log.Printf("AI服务调用失败: %v", err)
			aiChan <- "" // 发送空结果表示失败
		} else {
			aiChan <- aiResult
		}
	}()

	// 异步调用API服务
	go func() {
		recipes, err := h.recipeService.SearchByDishName(dishName)
		apiChan <- apiResult{recipes: recipes, err: err}
	}()

	// 等待AI结果
	aiResult := <-aiChan
	var aiError error
	if aiResult == "" {
		aiError = fmt.Errorf("AI服务不可用")
	}

	// 等待API结果
	apiRes := <-apiChan
	var apiError error
	if apiRes.err != nil {
		apiError = fmt.Errorf("API服务不可用")
	}

	// 构建最终结果
	var finalResult string
	var supplementaryData map[string]interface{}

	if aiError != nil && apiError != nil {
		// 两个服务都失败
		return "", nil, fmt.Errorf("所有服务都不可用")
	} else if aiError == nil && apiError != nil {
		// 只有AI服务可用
		finalResult = aiResult
		supplementaryData = map[string]interface{}{
			"ai_available":    true,
			"api_available":   false,
			"reference_count": 0,
		}
	} else if aiError != nil && apiError == nil {
		// 只有API服务可用
		recipesText := h.recipeService.FormatRecipesForAI(apiRes.recipes)
		finalResult = h.generateFallbackDishResult(dishName, recipesText)
		supplementaryData = map[string]interface{}{
			"ai_available":    false,
			"api_available":   true,
			"api_recipes":     apiRes.recipes,
			"reference_count": len(apiRes.recipes),
		}
	} else {
		// 两个服务都可用，整合结果
		finalResult = h.combineDishResults(aiResult, apiRes.recipes)
		supplementaryData = map[string]interface{}{
			"ai_available":    true,
			"api_available":   true,
			"api_recipes":     apiRes.recipes,
			"reference_count": len(apiRes.recipes),
		}
	}

	return finalResult, supplementaryData, nil
}

// generateFallbackIngredientResult 生成食材请求的备选结果
func (h *AgentHandler) generateFallbackIngredientResult(ingredients []string, recipesText string) string {
	ingredientsStr := strings.Join(ingredients, "、")
	return fmt.Sprintf(`# 食材分析与推荐

## 食材概述
您提供的食材：%s

这些食材搭配很有创意，可以制作出营养丰富、口感多样的菜品。

%s

## 简单制作建议

### 推荐制作方式
1. **清炒类**：将主要食材切块，大火快炒，保持食材营养和口感
2. **汤品类**：制作营养汤品，适合全家享用
3. **焖烧类**：慢火焖煮，让食材充分融合味道

### 烹饪小贴士
- 食材新鲜度是成功的关键
- 控制火候，避免过度烹饪
- 适当调味，突显食材本味

*注：当前显示基础推荐，如需个性化专业建议，请配置完整服务。*`, ingredientsStr, recipesText)
}

// generateFallbackDishResult 生成菜品请求的备选结果
func (h *AgentHandler) generateFallbackDishResult(dishName, recipesText string) string {
	return fmt.Sprintf(`# %s 制作指南

## 菜品介绍
%s是一道经典菜品，具有独特的风味和文化特色。

## 基础制作方法

### 食材准备
- 主要食材（根据菜品特点选择）
- 调料：盐、生抽、料酒等基础调料
- 辅料：姜、蒜、葱等增香材料

### 制作步骤
1. **准备工作**
   - 清洗和处理所有食材
   - 按需要改刀切配

2. **烹饪过程**
   - 热锅下油，控制油温
   - 按顺序下入食材
   - 适时调味，注意火候

3. **完成装盘**
   - 调整最终口味
   - 适当装饰，提升视觉效果

### 成功要点
- 食材处理要均匀一致
- 火候控制要精准
- 调味要层次分明

%s

*注：当前显示基础制作指南，如需详细专业指导，请配置完整服务。*`, dishName, dishName, recipesText)
}

// combineIngredientResults 整合食材分析结果
func (h *AgentHandler) combineIngredientResults(aiResult string, recipes []services.SpoonacularRecipe) string {
	recipesText := h.recipeService.FormatRecipesForAI(recipes)
	return aiResult + "\n\n" + "## API食谱参考\n\n" + recipesText
}

// combineDishResults 整合菜品详情结果
func (h *AgentHandler) combineDishResults(aiResult string, recipes []services.SpoonacularRecipe) string {
	recipesText := h.recipeService.FormatRecipesForAI(recipes)
	return aiResult + "\n\n" + "## 参考食谱信息\n\n" + recipesText
}