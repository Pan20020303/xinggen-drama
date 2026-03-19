# 🎬 星亘 Drama - AI Short Drama Production Platform

<div align="center">

**Full-stack AI Short Drama Automation Platform Based on Go + Vue3**

[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.x-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-CC%20BY--NC--SA%204.0-lightgrey.svg)](https://creativecommons.org/licenses/by-nc-sa/4.0/)

[Features](#features) • [Quick Start](#quick-start) • [Deployment](#deployment)

[简体中文](README-CN.md) | [English](README.md) | [日本語](README-JA.md)

</div>

---

## 📖 About

星亘 Drama is an AI-powered short drama production platform that automates the entire workflow from script generation, character design, storyboarding to video composition.

星亘短剧商业版地址：[星亘短剧商业版](https://drama.chatfire.site/shortvideo)

星亘小说生成：[星亘小说生成](https://marketing.chatfire.site/xinggen-novel/)

### 🎯 Core Features

- **🤖 AI-Driven**: Parse scripts using large language models to extract characters, scenes, and storyboards
- **🎨 Intelligent Creation**: AI-generated character portraits and scene backgrounds
- **📹 Video Generation**: Automatic storyboard video generation using text-to-video and image-to-video models
- **🔄 Complete Workflow**: End-to-end production workflow from idea to final video。

### 🛠️ Technical Architecture

Based on **DDD (Domain-Driven Design)** with clear layering:

```
├── API Layer (Gin HTTP)
├── Application Service Layer (Business Logic)
├── Domain Layer (Domain Models)
└── Infrastructure Layer (Database, External Services)
```

### 🎥 Demo Videos

Experience AI short drama generation:

<div align="center">

**Sample Work 1**

<video src="https://ffile.chatfire.site/cf/public/20260114094337396.mp4" controls width="640"></video>

**Sample Work 2**

<video src="https://ffile.chatfire.site/cf/public/fcede75e8aeafe22031dbf78f86285b8.mp4" controls width="640"></video>

[Watch Video 1](https://ffile.chatfire.site/cf/public/20260114094337396.mp4) | [Watch Video 2](https://ffile.chatfire.site/cf/public/fcede75e8aeafe22031dbf78f86285b8.mp4)

</div>

---

## ✨ Features

### 🎭 Character Management

- ✅ AI-generated character portraits
- ✅ Batch character generation
- ✅ Character image upload and management

### 🎬 Storyboard Production

- ✅ Automatic storyboard script generation
- ✅ Scene descriptions and shot design
- ✅ Storyboard image generation (text-to-image)
- ✅ Frame type selection (first frame/key frame/last frame/panel)

### 🎥 Video Generation

- ✅ Automatic image-to-video generation
- ✅ Video composition and editing
- ✅ Transition effects

### 📦 Asset Management

- ✅ Unified asset library management
- ✅ Local storage support
- ✅ Asset import/export
- ✅ Task progress tracking

---

## 🚀 Quick Start

> **TL;DR:** For local development, start MySQL and RabbitMQ with Docker first, verify `configs/config.yaml`, then run `go run main.go` and `cd web && npm run dev`.

### ✅ Definition of Done

Startup is complete when:

- [ ] `docker compose ps` shows `mysql` and `rabbitmq` healthy
- [ ] `go run main.go` starts without `database` or `rabbitmq` connection errors
- [ ] `http://localhost:5678/health` returns `{"status":"ok"}`
- [ ] `cd web && npm run dev` starts successfully
- [ ] Frontend opens at `http://localhost:3012`

### 📋 Prerequisites

| Software | Version | Required For |
| --- | --- | --- |
| **Go** | 1.23+ | Backend local startup |
| **Node.js** | 18+ | Frontend local startup |
| **npm** | 9+ | Frontend dependency management |
| **Docker / Docker Compose** | Latest | MySQL + RabbitMQ dependencies |
| **FFmpeg** | 4.0+ | Image/video/audio processing |

#### Installing FFmpeg

**macOS**

```bash
brew install ffmpeg
```

**Ubuntu / Debian**

```bash
sudo apt update
sudo apt install ffmpeg
```

**Windows**

Download from [FFmpeg Official Site](https://ffmpeg.org/download.html) and add it to `PATH`.

Verify:

```bash
ffmpeg -version
```

### 📥 Installation

```bash
git clone https://github.com/chatfire-AI/xinggen-drama.git
cd xinggen-drama

go mod download

cd web
npm install
cd ..
```

### ⚙️ Configuration

Create your local configuration:

```bash
cp configs/config.example.yaml configs/config.yaml
```

For **host-based local startup** (`go run main.go` on your machine), make sure `configs/config.yaml` uses host ports instead of Docker service names:

```yaml
server:
  host: "0.0.0.0"
  port: 5678
  cors_origins:
    - "http://localhost:3012"

database:
  type: "mysql"
  host: "localhost"
  port: 3306
  user: "xinggen"
  password: "xinggen123"
  database: "xinggen_drama"

mq:
  enabled: true
  url: "amqp://xinggen_rmq:XinggenRmq_2026_StrongPass@localhost:5672/"
  queue_prefix: "xinggen"
  consumer_enabled: true
  consumer_concurrency: 4
  prefetch_count: 8

storage:
  type: "local"
  local_path: "./data/storage"
  base_url: "http://localhost:5678/static"
```

For **full Docker deployment**, the application container uses internal service names from `docker-compose.yml`, so the effective values are:

- MySQL host: `mysql`
- RabbitMQ host: `rabbitmq`
- App port: `5678`

### 🎯 Local Development Startup

#### 1. Start infrastructure dependencies

```bash
docker compose up -d mysql rabbitmq
```

Verify:

```bash
docker compose ps
```

Expected ports:

- MySQL: `localhost:3306`
- RabbitMQ AMQP: `localhost:5672`
- RabbitMQ Management UI: `http://localhost:15672`

Default RabbitMQ credentials from `docker-compose.yml`:

- Username: `xinggen_rmq`
- Password: `XinggenRmq_2026_StrongPass`

#### 2. Start backend

```bash
go run main.go
```

Backend endpoints:

- Health check: `http://localhost:5678/health`
- API base: `http://localhost:5678/api/v1`
- Static files: `http://localhost:5678/static`

#### 3. Start frontend

```bash
cd web
npm run dev
```

Frontend URL:

- `http://localhost:3012`

The Vite dev server proxies API requests to the Go backend.

### 🧪 Single-Service Startup

If you want the backend to serve the built frontend:

```bash
cd web
npm run build
cd ..
go run main.go
```

Access:

- `http://localhost:5678`

### 🗄️ Database Initialization

Tables are created automatically on startup through GORM AutoMigrate.

If an older SQLite file already exists at `data/drama_generator.db`, the Docker startup flow attempts a one-time SQLite -> MySQL import and writes a migration marker file to avoid duplicate imports.

### 🛠️ Common Startup Problems

#### RabbitMQ authentication failed

Symptom:

```text
failed to create rabbitmq task bus: connect rabbitmq: Exception (403) Reason: "username or password not allowed"
```

Cause:

- The RabbitMQ container is running with a persisted data volume created using older credentials.

Fix:

```bash
docker compose stop rabbitmq
docker compose rm -f rabbitmq
docker volume rm huobao-drama_xinggen-rabbitmq || docker volume rm xinggen-rabbitmq
docker compose up -d rabbitmq
```

Then verify:

```bash
docker exec xinggen-rabbitmq rabbitmqctl list_users
```

#### Port `5678` already in use

Symptom:

```text
listen tcp :5678: bind: address already in use
```

Fix on Linux / WSL:

```bash
ss -ltnp | grep 5678
fuser -k 5678/tcp
```

#### Frontend cannot reach backend

Check:

- Backend is running on `localhost:5678`
- `server.cors_origins` includes `http://localhost:3012`
- `web/vite.config.ts` proxy settings are intact

#### SQLite write permission error

Symptom:

```text
attempt to write a readonly database
```

Fix:

```bash
mkdir -p data/storage
chmod -R 755 data
```

---

## 📦 Deployment

### ☁️ Cloud One-Click Deployment (Recommended 3080Ti)

👉 [优云智算，一键部署](https://www.compshare.cn/images/CaWEHpAA8t1H?referral_code=8hUJOaWz3YzG64FI2OlCiB&ytag=GPU_YY_YX_GitHub_xinggenai)

> ⚠️ **Note**: Please save your data to local storage promptly when using cloud deployment

---

### 🐳 Docker Deployment (Recommended)

#### Method 1: Docker Compose (Recommended)

The default Docker Compose setup starts three services:

- `xinggen-mysql`: MySQL 8 database
- `xinggen-rabbitmq`: RabbitMQ task queue
- `xinggen-drama`: application service

When upgrading from a previous SQLite-based deployment, the application container will automatically run a one-time SQLite -> MySQL import if it detects the legacy SQLite file in the mounted data volume.

#### 🚀 China Network Acceleration (Optional)

If you are in China, pulling Docker images and installing dependencies may be slow. You can speed up the build process by configuring mirror sources.

**Step 1: Create environment variable file**

```bash
cp .env.example .env
```

**Step 2: Edit `.env` file and uncomment the mirror sources you need**

```bash
# Enable Docker Hub mirror (recommended)
DOCKER_REGISTRY=docker.1ms.run/

# Enable npm mirror
NPM_REGISTRY=https://registry.npmmirror.com/

# Enable Go proxy
GO_PROXY=https://goproxy.cn,direct

# Enable Alpine mirror
ALPINE_MIRROR=mirrors.aliyun.com
```

**Step 3: Build with docker compose (required)**

```bash
docker compose build
```

> **Important Note**:
>
> - ⚠️ You must use `docker compose build` to automatically load mirror source configurations from the `.env` file
> - ❌ If using `docker build` command, you need to manually pass `--build-arg` parameters
> - ✅ Always recommended to use `docker compose build` for building

**Performance Comparison**:

| Operation        | Without Mirrors | With Mirrors |
| ---------------- | --------------- | ------------ |
| Pull base images | 5-30 minutes    | 1-5 minutes  |
| Install npm deps | May fail        | Fast success |
| Download Go deps | 5-10 minutes    | 30s-1 minute |

> **Note**: Users outside China should not configure mirror sources, use default settings.

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

#### Method 2: Docker Command

> **Note**: Linux users need to add `--add-host=host.docker.internal:host-gateway` to access host services

```bash
# Run from Docker Hub
docker run -d \
  --name xinggen-drama \
  -p 5678:5678 \
  -v $(pwd)/data:/app/data \
  --restart unless-stopped \
  xinggen/xinggen-drama:latest

# View logs
docker logs -f xinggen-drama
```

**Local Build** (optional):

```bash
docker build -t xinggen-drama:latest .
docker run -d --name xinggen-drama -p 5678:5678 -v $(pwd)/data:/app/data xinggen-drama:latest
```

**Docker Deployment Advantages:**

- ✅ Ready to use with default configuration
- ✅ Environment consistency, avoiding dependency issues
- ✅ One-click start, no need to install Go, Node.js, FFmpeg
- ✅ Easy to migrate and scale
- ✅ Automatic health checks and restarts
- ✅ Automatic file permission handling

#### 🔗 Accessing Host Services (Ollama/Local Models)

The container is configured to access host services using `http://host.docker.internal:PORT`.

**Configuration Steps:**

1. **Start service on host (listen on all interfaces)**

   ```bash
   export OLLAMA_HOST=0.0.0.0:11434 && ollama serve
   ```

2. **Frontend AI Service Configuration**
   - Base URL: `http://host.docker.internal:11434/v1`
   - Provider: `openai`
   - Model: `qwen2.5:latest`

---

### 🏭 Traditional Deployment

#### 1. Build

```bash
# 1. Build frontend
cd web
npm run build
cd ..

# 2. Compile backend
go build -o xinggen-drama .
```

Generated files:

- `xinggen-drama` - Backend executable
- `web/dist/` - Frontend static files (embedded in backend)

#### 2. Prepare Deployment Files

Files to upload to server:

```
xinggen-drama            # Backend executable
configs/config.yaml     # Configuration file
data/                   # Data directory (optional, auto-created on first run)
```

#### 3. Server Configuration

```bash
# Upload files to server
scp xinggen-drama user@server:/opt/xinggen-drama/
scp configs/config.yaml user@server:/opt/xinggen-drama/configs/

# SSH to server
ssh user@server

# Modify configuration file
cd /opt/xinggen-drama
vim configs/config.yaml
# Set mode to production
# Configure domain and storage path

# Create data directory and set permissions (Important!)
# Note: Replace YOUR_USER with actual user running the service (e.g., www-data, ubuntu, deploy)
sudo mkdir -p /opt/xinggen-drama/data/storage
sudo chown -R YOUR_USER:YOUR_USER /opt/xinggen-drama/data
sudo chmod -R 755 /opt/xinggen-drama/data

# Grant execute permission
chmod +x xinggen-drama

# Start service
./xinggen-drama
```

#### 4. Manage Service with systemd

Create service file `/etc/systemd/system/xinggen-drama.service`:

```ini
[Unit]
Description=星亘 Drama Service
After=network.target

[Service]
Type=simple
User=YOUR_USER
WorkingDirectory=/opt/xinggen-drama
ExecStart=/opt/xinggen-drama/xinggen-drama
Restart=on-failure
RestartSec=10

# Environment variables (optional)
# Environment="GIN_MODE=release"

[Install]
WantedBy=multi-user.target
```

Start service:

```bash
sudo systemctl daemon-reload
sudo systemctl enable xinggen-drama
sudo systemctl start xinggen-drama
sudo systemctl status xinggen-drama
```

**⚠️ Common Issue: SQLite Write Permission Error**

If you encounter `attempt to write a readonly database` error:

```bash
# 1. Check current user running the service
sudo systemctl status xinggen-drama | grep "Main PID"
ps aux | grep xinggen-drama

# 2. Fix permissions (replace YOUR_USER with actual username)
sudo chown -R YOUR_USER:YOUR_USER /opt/xinggen-drama/data
sudo chmod -R 755 /opt/xinggen-drama/data

# 3. Verify permissions
ls -la /opt/xinggen-drama/data
# Should show owner as the user running the service

# 4. Restart service
sudo systemctl restart xinggen-drama
```

**Reason:**

- SQLite requires write permission on both the database file **and** its directory
- Needs to create temporary files in the directory (e.g., `-wal`, `-journal`)
- **Key**: Ensure systemd `User` matches data directory owner

**Common Usernames:**

- Ubuntu/Debian: `www-data`, `ubuntu`
- CentOS/RHEL: `nginx`, `apache`
- Custom deployment: `deploy`, `app`, current logged-in user

#### 5. Nginx Reverse Proxy

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:5678;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    # Direct access to static files
    location /static/ {
        alias /opt/xinggen-drama/data/storage/;
    }
}
```

---

## 🎨 Tech Stack

### Backend

- **Language**: Go 1.23+
- **Web Framework**: Gin 1.9+
- **ORM**: GORM
- **Database**: SQLite
- **Logging**: Zap
- **Video Processing**: FFmpeg
- **AI Services**: OpenAI, Gemini, Doubao, etc.

### Frontend

- **Framework**: Vue 3.4+
- **Language**: TypeScript 5+
- **Build Tool**: Vite 5
- **UI Components**: Element Plus
- **CSS Framework**: TailwindCSS
- **State Management**: Pinia
- **Router**: Vue Router 4

### Development Tools

- **Package Management**: Go Modules, npm
- **Code Standards**: ESLint, Prettier
- **Version Control**: Git

---

## 📝 FAQ

### Q: How can Docker containers access Ollama on the host?

A: Use `http://host.docker.internal:11434/v1` as Base URL. Note two things:

1. Host Ollama needs to listen on `0.0.0.0`: `export OLLAMA_HOST=0.0.0.0:11434 && ollama serve`
2. Linux users using `docker run` need to add: `--add-host=host.docker.internal:host-gateway`

See: [DOCKER_HOST_ACCESS.md](docs/DOCKER_HOST_ACCESS.md)

### Q: FFmpeg not installed or not found?

A: Ensure FFmpeg is installed and in the PATH environment variable. Verify with `ffmpeg -version`.

### Q: Frontend cannot connect to backend API?

A: Check if backend is running and port is correct. In development mode, frontend proxy config is in `web/vite.config.ts`.

### Q: Database tables not created?

A: GORM automatically creates tables on first startup, check logs to confirm migration success.

---

## 📋 Changelog

### v1.0.5 (2026-02-06)

#### 🎨 Major Features

- **🎭 Global Style System**: Introduced comprehensive style selection support across the entire project. Users can now define custom visual styles at the drama level, which automatically applies to all AI-generated content including characters, scenes, and storyboards, ensuring consistent artistic direction throughout the production.

- **✂️ Nine-Grid Sequence Image Cropping**: Added cropping tool for action sequence images. Users can now extract individual frames from 3x3 grid layouts and designate them as first frames, last frames, or keyframes for video generation, providing greater flexibility in shot composition and continuity.

#### 🚀 Enhancements

- **📐 Optimized Action Sequence Grid**: Enhanced the visual quality and layout of nine-grid action sequence images with improved spacing, alignment, and frame transitions.

- **🔧 Manual Grid Assembly**: Introduced manual grid composition tools supporting 2x2 (four-grid), 2x3 (six-grid), and 3x3 (nine-grid) layouts, allowing users to create custom action sequences from individual frames.

- **🗑️ Content Management**: Added delete functionality for both generated images and videos, enabling better asset organization and storage management.

### v1.0.4 (2026-01-27)

#### 🚀 Major Updates

- Introduced local storage strategy for generated content caching, effectively mitigating external resource link expiration risks
- Implemented Base64 encoding for embedded reference image transmission
- Fixed issue where shot image prompt state was not reset when switching shots
- Fixed issue where video duration displayed as 0 when adding library videos
- Added scene migration to episodes

#### Historical Data Migration

- Added migration script for processing historical data. For detailed instructions, please refer to [MIGRATE_README.md](MIGRATE_README.md)

### v1.0.3 (2026-01-16)

#### 🚀 Major Updates

- Pure Go SQLite driver (`modernc.org/sqlite`), supports `CGO_ENABLED=0` cross-platform compilation
- Optimized concurrency performance (WAL mode), resolved "database is locked" errors
- Docker cross-platform support for `host.docker.internal` to access host services
- Streamlined documentation and deployment guides

### v1.0.2 (2026-01-14)

#### 🐛 Bug Fixes / 🔧 Improvements

- Fixed video generation API response parsing issues
- Added OpenAI Sora video endpoint configuration
- Optimized error handling and logging

---

## 🤝 Contributing

Issues and Pull Requests are welcome!

1. Fork this project
2. Create a feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

---

## API Configuration Site

Configure in 2 minutes: [API Aggregation Site](https://api.chatfire.site/models)

---

## 👨‍💻 About Us

**AI 星亘 - AI Studio Startup**

- 🏠 **Location**: Nanjing, China
- 🚀 **Status**: Startup in Progress
- 📧 **Email**: [18550175439@163.com](mailto:18550175439@163.com)
- 🐙 **GitHub**: [https://github.com/chatfire-AI/xinggen-drama](https://github.com/chatfire-AI/xinggen-drama)

> _"Let AI help us do more creative things"_

## Community Group

![Community Group](drama.png)

- Submit [Issue](../../issues)
- Email project maintainers

---

<div align="center">

**⭐ If this project helps you, please give it a Star!**

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=chatfire-AI/xinggen-drama&type=date&legend=top-left)](https://www.star-history.com/#chatfire-AI/xinggen-drama&type=date&legend=top-left)

Made with ❤️ by 星亘 Team

</div>
