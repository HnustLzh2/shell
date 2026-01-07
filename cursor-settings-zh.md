## Cursor 设置中文说明

本文件简要说明 Cursor 主要设置页面及常用选项的含义，帮助你看懂英文界面。对应左侧栏中的：`General`、`Agents`、`Tab`、`Models`、`Tools & MCP`、`Rules and Commands`、`Indexing & Docs`、`Network`、`Beta`、`Docs` 等。

---

## 一、General（通用）

### 1. 账户与订阅

- **Manage Account**：管理账号和账单（在浏览器中打开账号页面）。
- **Upgrade to Ultra**：升级到 Ultra 套餐，获得更高调用额度、并行 Agent 等。

### 2. Preferences（偏好）

- **Default Layout**：默认布局  
  - **Agent**：侧重右侧 Agent/聊天面板。  
  - **Editor**：侧重代码编辑器区域。

- **Editor Settings**：编辑器设置  
  - 字体、字号、主题（Theme）、行号、缩进（Tab / 空格）、自动保存（Auto Save）、Minimap 等。

- **Keyboard Shortcuts**：键盘快捷键  
  - 搜索、修改、删除或新增快捷键绑定，通常也可以恢复默认（Reset）。

- **Import Settings from VS Code**：从 VS Code 导入设置  
  - 导入 VS Code 的主题、字体、键位绑定和部分扩展配置，让 Cursor 体验更接近你原来的 VS Code。

- **Reset “Don’t Ask Again” Dialogs**：重置“不再提示”的对话框  
  - 清除之前勾选“Don’t ask again”的记忆，让这些确认弹窗重新出现。

---

## 二、Agents（智能助手）

- **全局 Agent 行为**  
  - 是否默认让 Agent 读取整个工作区上下文。  
  - 默认回答语言、风格等（也可以在 Rules 中设定，如：`Always respond in 中文.`）。  
  - 是否允许 Agent 自动调用工具（运行命令、编辑文件等）。

- **Agent 预设 / Profiles（如果有）**  
  - 为不同用途建立不同 Agent 配置：解释代码、重构、写文档、学习等，并可在界面中切换。

---

## 三、Tab（标签页 / 布局）

- **Tab 行为**  
  - 单击 / 双击 打开文件；是否启用“预览标签页”（Preview Tab）等。

- **布局相关**  
  - Agent 面板与 Editor 面板是左右还是上下分布，是否在新标签中打开 Agent 对话等。

---

## 四、Models（模型）

- **默认模型选择**  
  - 为不同功能指定模型：  
    - 代码补全（Tab 补全）使用的模型。  
    - Chat / Agent 对话使用的模型。  
    - 大上下文分析使用的模型。

- **自定义 API Key（若支持）**  
  - 配置自己的 OpenAI / Anthropic 等 Key，用于使用自有配额或更多模型。

- **回退策略**  
  - 主模型不可用时是否自动切换到备用模型等。

---

## 五、Tools & MCP（工具与 MCP）

- **内置工具**  
  - 控制 Cursor 可调用的系统工具：  
    - 运行终端命令、读写文件、搜索代码、Git 操作等。  
  - 可以为某些危险操作限制权限或关闭自动执行。

- **MCP（Model Context Protocol）工具**  
  - 注册来自服务器或第三方的 MCP 工具（搜索、数据库、内部服务等）。  
  - 为每个 MCP 配置连接信息、访问 Token，并决定是否允许 Agent 使用。

---

## 六、Rules and Commands（规则与命令）

### 1. Project Rules（项目规则）

- 为“当前项目”定义专属对话规则，只在这个仓库生效。  
- 示例：  
  - “This is a Go project, use go test ./... for tests.”  
  - “Always respond in 中文.”  
  - “Follow this repository’s coding style.”

### 2. Project Commands（项目命令）

- 为当前项目定义可复用命令（类似脚本快捷方式）：  
  - 命令名：如 `Run API Tests`。  
  - 实际命令：如 `go test ./app/...`。  
- 可在命令面板或右下角 Command 区快速调用。

### 3. User Commands（用户命令）

- 对所有项目通用的命令。  
- 适合放日常都会用到的通用操作，如格式化、常用脚本等。

---

## 七、Indexing & Docs（索引与文档）

- **代码索引（Indexing）**  
  - 决定 Cursor 是否/如何索引当前仓库：  
    - 索引后，搜索和 Agent 参考代码会更快更准。  
  - 可配置：  
    - 忽略哪些目录（如依赖、生成文件）。  
    - 重新索引、清除索引缓存等。

- **Docs（文档源）**  
  - 为项目添加要索引的文档（如 Markdown 笔记、设计文档等）。  
  - 索引完成后，Agent 回答时可以引用这些文档内容。

---

## 八、Network（网络）

- **代理（Proxy）**  
  - 配置 HTTP / HTTPS 代理地址，适用于：  
    - 公司内网需要代理才能访问外网。  
    - 访问模型服务需要特殊网络环境。

- **证书 / 信任设置（视版本而定）**  
  - 为自签名证书或内部 CA 配置信任策略。

---

## 九、Beta（实验功能）

- **实验特性开关**  
  - 试用正在测试中的新功能：  
    - 新 UI、新模型集成、新工具行为等。  
  - 若遇到不稳定，可关闭恢复为稳定版本行为。

---

## 十、Docs（文档）

- **官方文档入口**  
  - 打开 Cursor 官方文档网站，包含：  
    - 快速上手、配置说明、FAQ、更新日志等。

---

## 建议用法

- 想尽量“中文化体验”：  
  - 在 `Rules and Commands → Project Rules` 或用户规则中添加：`Always respond in 中文.`  
  - 这样虽然 UI 还是英文，但智能助手回答、说明和注释都会使用中文。

- 如果碰到设置项看不懂：  
  - 可以把页面截图或英文选项复制给我，我可以基于你当前版本的实际文案做更细的解释和翻译。


