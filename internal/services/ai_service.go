package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// AIService AI服务结构
type AIService struct {
	apiKey     string
	baseURL    string
	model      string
}

// DeepSeekAPIRequest DeepSeek API请求结构
type DeepSeekAPIRequest struct {
	Model    string         `json:"model"`
	Messages []ChatMessage  `json:"messages"`
	Stream   bool           `json:"stream"`
}

// ChatMessage 聊天消息结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekAPIResponse DeepSeek API响应结构
type DeepSeekAPIResponse struct {
	ID      string           `json:"id"`
	Object  string           `json:"object"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Choices []APIChoice      `json:"choices"`
	Usage   APIUsage         `json:"usage"`
}

// APIChoice API选择项
type APIChoice struct {
	Index   int         `json:"index"`
	Message ChatMessage `json:"message"`
	Finish  string      `json:"finish_reason"`
}

// APIUsage API使用情况
type APIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewAIService 创建AI服务实例
func NewAIService() *AIService {
	return &AIService{
		apiKey:  os.Getenv("DEEPSEEK_API_KEY"),
		baseURL: "https://api.deepseek.com/v1/chat/completions",
		model:   "deepseek-chat",
	}
}

// AnalyzeIngredients 根据食材分析菜品
func (s *AIService) AnalyzeIngredients(ingredients []string) (string, error) {
	if s.apiKey == "" {
		return s.generateDefaultRecipe(ingredients), nil
	}

	prompt := s.buildIngredientsPrompt(ingredients)
	return s.callDeepSeekAPI(prompt)
}

// GetDishDetails 获取菜品详细制作方法
func (s *AIService) GetDishDetails(dishName string) (string, error) {
	if s.apiKey == "" {
		return s.generateDefaultDishDetails(dishName), nil
	}

	prompt := s.buildDishPrompt(dishName)
	return s.callDeepSeekAPI(prompt)
}

// buildIngredientsPrompt 构建食材分析prompt
func (s *AIService) buildIngredientsPrompt(ingredients []string) string {
	ingredientsText := ""
	for i, ingredient := range ingredients {
		if i > 0 {
			ingredientsText += "、"
		}
		ingredientsText += ingredient
	}

	return fmt.Sprintf(`你是一位专业的厨师和营养师。请根据用户提供的食材，给出专业、实用的烹饪建议。

用户提供的食材：%s

请按照以下结构提供分析：

## 🍳 推荐菜品

### 1. [主要推荐菜品名称]
**简介**：[简要介绍菜品特点和风味]
**匹配度**：⭐️⭐️⭐️⭐️⭐️ (基于现有食材的匹配程度)
**所需完整食材**：
- ✅ 已有：%s
- 🔶 需要补充：[列出需要额外购买的食材]
**烹饪步骤**：
1. [第一步详细说明]
2. [第二步详细说明]
3. [第三步详细说明]
**烹饪技巧**：[专业小贴士]
**预计时间**：[准备时间 + 烹饪时间]
**难度等级**：★☆☆☆☆ (简单) / ★★☆☆☆ (中等) / ★★★☆☆ (有一定难度)

### 2. [次要推荐菜品名称]
[同样结构...]

### 3. [创意菜品名称]
[同样结构...]

## 📊 营养分析
- **主要营养**：[蛋白质、维生素等分析]
- **适合人群**：[适合什么人群食用]
- **健康建议**：[饮食建议]

## 💡 其他可能组合
- [食材的其他简单搭配建议]

请确保：
1. 推荐真实可行的家常菜品
2. 步骤清晰易懂，适合家庭厨房操作
3. 标注清楚需要额外购买的食材
4. 提供实用的烹饪技巧
5. 用中文回复，语气亲切专业`, ingredientsText, ingredientsText)
}

// buildDishPrompt 构建菜品详情prompt
func (s *AIService) buildDishPrompt(dishName string) string {
	return fmt.Sprintf(`你是一位经验丰富的专业厨师。请为用户提供"%s"的完整、详细的烹饪教程。

请按照以下结构提供信息：

## 🥘 菜品详情
**菜系**：[所属菜系，如川菜、粤菜等]
**口味特点**：[麻辣、清淡、酸甜等]
**文化背景**：[简要的起源或文化背景]

## 🛒 食材清单
### 主料
- [主要食材1]：[用量] + [挑选技巧]
- [主要食材2]：[用量] + [挑选技巧]

### 辅料/调料
- [调料1]：[用量] + [作用说明]
- [调料2]：[用量] + [作用说明]

## 👨‍🍳 详细步骤
### 准备阶段 (约X分钟)
1. **[预处理]**：[食材处理详细步骤]
2. **[切配]**：[刀工要求和切配方法]
3. **[调料准备]**：[调料调配方法]

### 烹饪阶段 (约Y分钟)
1. **第一步**：[火候] + [操作细节] + [时长]
   - 💡 技巧：[关键技巧说明]
2. **第二步**：[火候] + [操作细节] + [时长]
   - ⚠️ 注意：[容易出错的地方]
3. **第三步**：[火候] + [操作细节] + [时长]

## 🎯 成功关键
- **火候控制**：[具体火候要求]
- **时机把握**：[关键时间点]
- **调味顺序**：[调料添加顺序的重要性]

## 📈 难度分析
- **准备难度**：★☆☆☆☆
- **烹饪难度**：★☆☆☆☆
- **总耗时**：约X分钟

## 🌶️ 口味调整
- **更辣**：[调整方法]
- **更清淡**：[调整方法]
- **素食版**：[替代方案]

## 🍽️ 搭配建议
- **主食搭配**：[推荐搭配的主食]
- **配菜推荐**：[推荐的配菜]
- **饮品搭配**：[适合的饮品]

请确保步骤详细、准确，适合家庭厨房操作，包含专业厨师的实用技巧。回复时保持段落紧凑，减少不必要的换行。`, dishName)
}

// callDeepSeekAPI 调用DeepSeek API
func (s *AIService) callDeepSeekAPI(prompt string) (string, error) {
	requestBody := DeepSeekAPIRequest{
		Model: s.model,
		Messages: []ChatMessage{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Stream: false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %v", err)
	}

	req, err := http.NewRequest("POST", s.baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API调用失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	var apiResp DeepSeekAPIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("解析响应失败: %v", err)
	}

	if len(apiResp.Choices) == 0 {
		return "", fmt.Errorf("API返回空响应")
	}

	return apiResp.Choices[0].Message.Content, nil
}

// generateDefaultRecipe 生成默认食谱（当API不可用时）
func (s *AIService) generateDefaultRecipe(ingredients []string) string {
	ingredientsText := ""
	for i, ingredient := range ingredients {
		if i > 0 {
			ingredientsText += "、"
		}
		ingredientsText += ingredient
	}

	return fmt.Sprintf(`# 食材分析与菜品推荐

## 食材分析
您提供的食材：%s

这些食材搭配合理，可以制作出营养均衡的美味菜品。

## 推荐菜品

### 菜品1：家常炒菜
**制作难度：** 简单
**预计时间：** 20分钟

**所需食材：**
- 主料：%s
- 调料：盐、生抽、食用油适量

**制作步骤：**
1. 将所有食材洗净切好备用
2. 热锅下油，爆香配料
3. 下入主料大火翻炒
4. 调味炒匀即可出锅

### 菜品2：营养汤品
**制作难度：** 简单
**预计时间：** 30分钟

**制作步骤：**
1. 食材洗净切成适当大小
2. 锅中加水烧开
3. 依次下入食材煮制
4. 调味后撒上葱花即可

## 烹饪小贴士
- 食材新鲜度很重要
- 火候控制要适宜
- 调味要适量

*注：当前使用默认推荐，如需更详细的个性化建议，请配置AI服务。*`, ingredientsText, ingredientsText)
}

// generateDefaultDishDetails 生成默认菜品详情（当API不可用时）
func (s *AIService) generateDefaultDishDetails(dishName string) string {
	return fmt.Sprintf(`# %s 制作指南

## 菜品介绍
%s是一道经典的菜品，深受人们喜爱。这道菜口感丰富，营养均衡，适合家庭制作。

## 食材清单
### 主料
- 主要食材（根据菜品特点）
- 适量配菜

### 调料
- 食用油：适量
- 盐：适量
- 生抽：2勺
- 料酒：1勺
- 其他调料适量

## 制作步骤

### 1. 准备工作
- 将所有食材洗净
- 按需要改刀处理
- 备好所有调料

### 2. 烹饪过程
1. 热锅下油，油温适中
2. 按顺序下入食材
3. 控制火候，适时翻炒
4. 调味炒匀

### 3. 出锅装盘
- 调整最终味道
- 装盘装饰即可

## 成功要点
- 食材处理要均匀
- 火候控制要精准
- 调味要层次分明

*注：当前使用默认制作指南，如需专业指导，请配置AI服务。*`, dishName, dishName)
}