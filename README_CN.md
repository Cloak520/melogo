# MeloGo

一个使用 Go 构建的自托管音乐流媒体服务器，具有自动音乐库扫描、播放列表管理和基于 Web 的音乐播放界面。

## 功能特性

- **自动音乐库扫描**: 自动扫描您的音乐目录并提取元数据（标题、艺术家、专辑、时长）
- **基于 Web 的音乐播放器**: 通过现代化的 Web 界面流式播放您的音乐收藏
- **用户认证**: 基于 JWT 的安全登录和注册系统
- **播放列表管理**: 创建、编辑和管理自定义播放列表
- **收藏系统**: 标记您最喜欢的歌曲以便轻松访问
- **歌词支持**: 自动歌词抓取和显示
- **封面艺术**: 自动封面艺术提取和显示
- **搜索功能**: 按标题、艺术家或专辑搜索您的音乐库
- **多语言支持**: 支持多种语言的国际化
- **管理面板**: 用于管理音乐库的管理工具
- **Docker 支持**: 使用 Docker 和 Docker Compose 轻松部署

## 技术栈

- **后端**: Go (Gin 框架)
- **数据库**: SQLite
- **前端**: HTML, CSS, JavaScript
- **认证**: JWT (JSON Web Tokens)
- **音乐元数据**: go-taglib
- **国际化**: go-i18n
- **模板引擎**: Go 的 html/template

## 安装

### 前置要求

- Go 1.25 或更高版本
- Git

### 快速开始

1. 克隆仓库：
   ```bash
   git clone https://github.com/your-username/MeloGo.git
   cd MeloGo
   ```

2. 安装依赖：
   ```bash
   go mod tidy
   ```

3. 构建应用程序：
   ```bash
   go build -o melogo
   ```

4. 通过编辑 `.env` 文件配置应用程序：
   ```bash
   # 编辑 .env 来配置您的设置
   cp .env.example .env
   ```

5. 运行应用程序：
   ```bash
   ./melogo
   ```
   或使用自定义 .env 文件：
   ```bash
   ./melogo -f /path/to/your/.env
   ```

## 配置

应用程序可以使用 `.env` 文件中的环境变量进行配置：

- `SERVER_HOST`: 服务器主机 (默认: localhost)
- `SERVER_PORT`: 服务器端口 (默认: 8080)
- `SERVER_DEBUG`: 启用调试模式 (默认: false)
- `DATABASE_PATH`: SQLite 数据库文件路径 (默认: ./data/melogo.db)
- `MUSIC_DIRECTORY`: 包含音乐文件的目录 (默认: ./music)
- `MUSIC_SCAN_INTERVAL`: 音乐扫描间隔（分钟）(默认: 5)
- `ALLOW_REGISTRATION`: 允许用户注册 (默认: true)
- `JWT_SECRET`: JWT 密钥 (生产环境中请更改!)
- `LYRICS_API_URL`: 歌词抓取的 API URL (默认: https://api.lrc.cx)

## 使用

1. 将您的音乐文件放在配置的音乐目录中
2. 启动服务器 - 它将自动扫描并索引您的音乐
3. 在 `http://localhost:8080` 访问 Web 界面
4. 注册账户或登录
5. 浏览、搜索和播放您的音乐收藏
6. 创建播放列表并标记收藏

## Docker 部署

MeloGo 可以使用 Docker 轻松部署：

```bash
# 拉取并运行官方镜像
docker run -d \
  --name melogo \
  -p 8080:8080 \
  -v ./data:/home/melogo/data \
  -v ./music:/home/melogo/music \
  -e SERVER_HOST=0.0.0.0 \
  -e JWT_SECRET=your-secret-key \
  soulcloak/melogo:latest
```

或使用 Docker Compose（推荐）：

```yaml
version: '3.8'

services:
  melogo:
    image: soulcloak/melogo:latest
    container_name: melogo
    environment:
      - SERVER_HOST=0.0.0.0
      - SERVER_PORT=8080
      - DATABASE_PATH=./data/melogo.db
      - MUSIC_DIRECTORY=./music
      - JWT_SECRET=your-secret-key
      - ALLOW_REGISTRATION=true
      - SERVER_DEBUG=false
      - LYRICS_API_URL=https://api.lrc.cx
    ports:
      - "8080:8080"
    volumes:
      - ./data:/home/melogo/data
      - ./music:/home/melogo/music
    restart: unless-stopped
```

## API 端点

MeloGo 在 `/api/v1` 提供 RESTful API：

- `POST /api/v1/register` - 用户注册
- `POST /api/v1/login` - 用户登录
- `POST /api/v1/logout` - 用户登出
- `GET /api/v1/user/profile` - 获取用户资料
- `PUT /api/v1/user/profile` - 更新用户资料
- `GET /api/v1/songs` - 列出所有歌曲
- `GET /api/v1/songs/:id` - 获取歌曲详情
- `GET /api/v1/songs/:id/stream` - 流式播放歌曲音频
- `GET /api/v1/songs/:id/lyrics` - 获取歌曲歌词
- `GET /api/v1/songs/:id/cover` - 获取歌曲封面图片
- `GET /api/v1/playlists` - 列出用户播放列表
- `POST /api/v1/playlists` - 创建播放列表
- `PUT /api/v1/playlists/:id` - 更新播放列表
- `DELETE /api/v1/playlists/:id` - 删除播放列表
- `GET /api/v1/playlists/:id/detail` - 获取播放列表详情
- `POST /api/v1/playlists/:id/songs` - 将歌曲添加到播放列表
- `DELETE /api/v1/playlists/:id/songs/:song_id` - 从播放列表中移除歌曲
- `GET /api/v1/favorites` - 列出收藏的歌曲
- `POST /api/v1/favorites` - 将歌曲添加到收藏
- `DELETE /api/v1/favorites/:song_id` - 从收藏中移除歌曲
- `GET /api/v1/search` - 搜索歌曲
- `GET /api/v1/search/history` - 获取搜索历史

## 开发

要为 MeloGo 做出贡献：

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 进行更改
4. 提交更改 (`git commit -m 'Add some amazing feature'`)
5. 推送到分支 (`git push origin feature/amazing-feature`)
6. 打开 Pull Request

### 在开发模式下运行

```bash
# 通过在 .env 中设置 SERVER_DEBUG=true 启用调试模式
go run main.go
```

## 贡献

欢迎贡献！请随时提交 Pull Request。对于重大更改，请先开一个 issue 来讨论您想要更改的内容。

## 许可证

此项目根据 MIT 许可证授权 - 请参阅 [LICENSE](LICENSE) 文件了解详情。