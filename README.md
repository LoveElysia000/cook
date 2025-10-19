# 智能食谱助手

一个基于AI的智能食谱推荐系统，支持根据食材推荐菜品或根据菜名提供详细制作方法。

## 功能特性

### 🍳 核心功能
- **按食材搜索**: 输入食材，AI智能分析并推荐适合的菜品
- **按菜名搜索**: 输入菜名，获取详细的制作指导和技巧
- **智能分析**: 基于DeepSeek AI的专业食谱分析
- **现代化界面**: 支持桌面和移动设备的优雅Web界面，丰富动画效果

### 🎨 v2.2.0界面更新亮点
- **现代化设计**: 采用青绿色和珊瑚橙渐变配色系统
- **丰富动画**: 标题发光、图标弹跳、流光扫过等40+种动画效果
- **微交互**: 按钮涟漪效果、输入框3D浮动、加载条流光动画
- **响应式优化**: 针对平板、手机、小屏设备的专门适配
- **无障碍支持**: 暗色主题自动适配、减少动画选项、高对比度支持

### 🚀 技术特性
- **并行处理**: AI分析和API数据查询并行执行
- **智能缓存**: 内置缓存机制提升响应速度
- **动态翻译**: AI驱动的中英文食材/菜品翻译系统
- **容错设计**: 服务降级确保系统可用性
- **模块化架构**: 清晰的代码结构便于维护和扩展

## 技术栈

| 组件 | 技术选择 |
|------|----------|
| 后端 | Go + Gin框架 |
| 前端 | HTML5 + CSS3 + JavaScript + Bootstrap 5 |
| AI服务 | DeepSeek API |
| 数据源 | Spoonacular API (可选) |
| 缓存 | 内存缓存 |
| 动画系统 | CSS3 Animation + JavaScript |

## 快速开始

### 环境要求
- Go 1.19+
- 现代浏览器 (Chrome, Firefox, Safari, Edge)

### 安装和运行

1. **克隆项目**
   ```bash
   git clone <项目地址>
   cd cook-agent
   ```

2. **安装依赖**
   ```bash
   go mod tidy
   ```

3. **配置环境变量**
   ```bash
   # 复制配置文件模板
   cp .env.example .env

   # 编辑配置文件，填入API密钥
   nano .env
   ```

4. **获取API密钥** (可选)
   - **DeepSeek API**: 访问 [https://platform.deepseek.com/](https://platform.deepseek.com/) 注册获取
   - **Spoonacular API**: 访问 [https://spoonacular.com/food-api](https://spoonacular.com/food-api) 注册获取

5. **启动服务**
   ```bash
   go run main.go
   ```

6. **访问应用**
   ```
   http://localhost:8080
   ```

## 配置说明

### 环境变量

| 变量名 | 必需 | 默认值 | 说明 |
|--------|------|--------|------|
| `PORT` | 否 | 8080 | 服务端口 |
| `GIN_MODE` | 否 | release | 运行模式 (debug/release) |
| `DEEPSEEK_API_KEY` | 否 | - | DeepSeek API密钥 |
| `SPOONACULAR_API_KEY` | 否 | - | Spoonacular API密钥 |

### API密钥说明

- 即使不配置API密钥，应用也能正常运行，但功能会受限
- 无DeepSeek API时: 使用默认的食谱推荐逻辑
- 无Spoonacular API时: 仅使用AI分析和本地逻辑

## 使用指南

### 按食材搜索
1. 在"按食材搜索"标签页下输入食材名称
2. 按回车键添加到食材列表
3. 点击"智能分析"获取推荐菜品

### 按菜名搜索
1. 切换到"按菜名搜索"标签页
2. 输入菜品名称
3. 点击"智能分析"获取制作指导

### 键盘快捷键
- `Ctrl/Cmd + Enter`: 执行搜索
- `Escape`: 清空输入
- `Enter`: 添加食材/执行搜索

## 项目结构

```
cook-agent/
├── main.go                    # 应用入口
├── go.mod                     # Go模块定义
├── go.sum                     # 依赖锁定文件
├── .env                       # 环境变量配置
├── .env.example              # 配置文件模板
├── README.md                 # 项目说明
├── CHANGELOG.md              # 版本更新日志
├── templates/                # HTML模板
│   └── index.html
├── static/                   # 静态资源
│   ├── css/
│   │   └── style.css         # 1062行现代化CSS样式
│   └── js/
│       └── app.js
└── internal/                 # 内部模块
    ├── handlers/             # HTTP处理器
    │   ├── handlers.go
    │   └── agent_handler.go
    └── services/             # 业务服务
        ├── ai_service.go
        ├── recipe_service.go
        └── translation_service.go  # AI驱动的翻译服务
```

### 📊 代码统计 (v2.2.0)
- **总代码行数**: 1062+ 行
- **CSS样式**: 1062行 (包含40+动画效果)
- **Go代码**: 500+ 行
- **HTML模板**: 194行
- **总文件数**: 11个核心文件

## API文档

### POST /api/recipes

请求格式:
```json
{
  "ingredients": ["鸡蛋", "西红柿"],
  "dishName": "西红柿炒鸡蛋",
  "queryType": "ingredients" // 或 "dish"
}
```

响应格式:
```json
{
  "result": "AI生成的食谱内容",
  "type": "ingredients",
  "timestamp": "2024-01-01T10:00:00Z",
  "supplementaryData": {
    "ai_available": true,
    "api_available": true,
    "api_recipes": [...],
    "nutrition_tips": "..."
  },
  "success": true
}
```

### GET /api/health

健康检查接口，返回服务状态。

## 部署选项

### Docker部署
```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o recipe-agent main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/recipe-agent .
COPY --from=builder /app/templates templates
COPY --from=builder /app/static static
EXPOSE 8080
CMD ["./recipe-agent"]
```

### 生产环境配置
```bash
# 使用生产环境配置
export GIN_MODE=release
export PORT=8080
export DEEPSEEK_API_KEY="your_production_key"
export SPOONACULAR_API_KEY="your_production_key"

# 启动服务
./recipe-agent
```

## 性能优化

### 缓存策略
- 食材分析结果缓存30分钟
- 菜品详情缓存60分钟
- 自动清理过期缓存

### 并发处理
- AI分析和API查询并行执行
- 超时控制防止阻塞
- 容错机制确保服务可用

## 开发指南

### 添加新的AI服务
1. 在 `internal/services/ai_service.go` 中添加新方法
2. 更新处理器调用逻辑
3. 配置相应的环境变量

### 扩展数据源
1. 在 `internal/services/recipe_service.go` 中添加新API集成
2. 实现相应的数据格式化方法
3. 更新缓存策略

## 故障排除

### 常见问题

1. **服务启动失败**
   - 检查端口是否被占用
   - 验证Go环境是否正确安装

2. **AI服务调用失败**
   - 验证DeepSeek API密钥是否正确
   - 检查网络连接

3. **前端样式异常**
   - 确认静态资源路径正确
   - 检查浏览器控制台错误

### 日志查看
应用运行时会输出详细日志，可通过以下方式查看：
```bash
# 开发模式
go run main.go

# 生产模式(后台运行)
nohup ./recipe-agent > app.log 2>&1 &
```

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 版本历史

### v2.2.0 - 前端界面全面升级 (2025-10-19)
- 🎨 全新的现代化UI设计，青绿色和珊瑚橙渐变配色
- ✨ 40+种动画效果：标题发光、图标弹跳、流光扫过等
- 🎯 微交互优化：按钮涟漪、输入框3D浮动、加载条流光
- 📱 完善的响应式设计：平板/手机/小屏设备专门适配
- ♿ 无障碍支持：暗色主题、减少动画、高对比度主题
- 🎭 CSS变量系统：支持主题定制和全局样式管理

### v2.1.0 - 动态翻译功能升级 (2025-10-18)
- 🤖 AI驱动的动态翻译系统替代静态映射表
- ⚡ 24小时翻译缓存，避免重复API调用
- 🌍 支持无限扩展的中文食材/菜品名称
- 🔧 三层翻译策略：高频映射→AI翻译→降级匹配

### v2.0.0 - 初始版本
- 🍳 基础食谱搜索功能
- 🤖 AI烹饪建议
- 📊 Spoonacular API集成
- 📝 中英文静态映射表

查看完整的[更新日志](CHANGELOG.md)了解详细信息。

## 致谢

- [DeepSeek](https://www.deepseek.com/) - 提供强大的AI能力
- [Spoonacular](https://spoonacular.com/) - 提供丰富的食谱数据
- [Gin](https://gin-gonic.com/) - 高性能的Go Web框架
- [Bootstrap](https://getbootstrap.com/) - 现代化的前端UI框架

## 联系方式

如有问题或建议，请通过以下方式联系:
- 🐛 提交 Issue
- 📧 发送邮件到项目维护者
- ⭐ 给项目一个Star支持我们的工作

---

享受烹饪的乐趣！ 🍽️✨