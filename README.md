# 🛡️ TierLog — Enterprise Intelligent Thesis Supervision Platform

[![Tech Stack](https://img.shields.io/badge/Stack-Go%20%7C%20Laravel%20%7C%20Vue%203%20%7C%20Inertia-blue?style=for-the-badge)](https://github.com/)
[![Backend](https://img.shields.io/badge/Backend-Go%201.22%2B%20(Gin)-00ADD8?style=for-the-badge&logo=go)](https://golang.org/)
[![Frontend](https://img.shields.io/badge/Frontend-Laravel%2011%20%7C%20Vue%203-FF2D20?style=for-the-badge&logo=laravel)](https://laravel.com/)
[![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=for-the-badge&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-Enterprise%20Proprietary-darkgreen?style=for-the-badge)](https://github.com/)

**TierLog** is an enterprise-grade, high-performance, real-time thesis supervision (*bimbingan tugas akhir*) and revision tracking platform. Designed for modern academic institutions, TierLog combines a highly concurrent micro-service architecture powered by a **Go (Gin Gonic) API Gateway** with a reactive, desktop-class **Laravel 11 + Inertia.js (Vue 3, TypeScript, Tailwind CSS v4) Frontend**.

By integrating advanced multimodal AI engines—including Groq Whisper speech-to-text, NVIDIA NIM, OpenAI, Anthropic, and Gemini Vision—TierLog automates consultation transcribing, handwritten notes recognition (OCR), track-changes parsing, and intelligent revision classification while enforcing strict academic guardrails against AI hallucinations.

---

## 📋 Table of Contents

- [1. Enterprise System Architecture](#1-enterprise-system-architecture)
- [2. Key Features & Technological Innovations](#2-key-features--technological-innovations)
  - [2.1 AI Dosen Persona Guardrails (Anti-Hallucination Engine)](#21-ai-dosen-persona-guardrails-anti-hallucination-engine)
  - [2.2 Groq Whisper Audio Slicing Engine (STT)](#22-groq-whisper-audio-slicing-engine-stt)
  - [2.3 Multimodal Vision OCR & Track Changes Parsing](#23-multimodal-vision-ocr--track-changes-parsing)
  - [2.4 Resilient State-Machine JSON Sanitizer](#24-resilient-state-machine-json-sanitizer)
  - [2.5 Provider-Agnostic LLM Gateway & Dual Key Architecture](#25-provider-agnostic-llm-gateway--dual-key-architecture)
  - [2.6 Realtime WebSocket Hub & Event Broadcasting](#26-realtime-websocket-hub--event-broadcasting)
  - [2.7 Widescreen Desktop Workspace Engine (1600px)](#27-widescreen-desktop-workspace-engine-1600px)
- [3. Project Directory Structure](#3-project-directory-structure)
- [4. Database Architecture & Data Models](#4-database-architecture--data-models)
- [5. API & WebSocket Contract Specifications](#5-api--websocket-contract-specifications)
- [6. Installation & Environment Configuration](#6-installation--environment-configuration)
- [7. Production Deployment & Containerization](#7-production-deployment--containerization)
- [8. Enterprise Security & Reliability Standard](#8-enterprise-security--reliability-standard)

---

## 1. Enterprise System Architecture

TierLog utilizes a decoupled dual-engine architecture optimized for low latency, sub-second WebSocket broadcasting, high concurrency, and secure data handling.

```mermaid
flowchart TB
    subgraph Client [Client Tier — Single Page Application]
        direction TB
        UI[Vue 3 Components & Floating Windows] <--> Stores[Pinia State Stores & Inertia Router]
        UI <--> Widescreen[Dynamic Workspace Engine - 1600px]
    end

    subgraph WebServer [Frontend Web Service — Laravel 11: Port 8000]
        direction TB
        AuthProxy[Session Routing & SSR Proxy]
        Inertia[Inertia.js Server Renderer]
        SQLite[(Local SQLite App Cache)]
        
        AuthProxy --> Inertia
        Inertia --> SQLite
    end

    subgraph APIServer [Core API Backend Service — Go 1.22 (Gin): Port 8080]
        direction TB
        Gin[Gin Router & Middleware Limiter]
        GORM[GORM ORM Core Engine]
        WSHub[Realtime WebSocket Room Hub]
        AICtrl[AI Dispatcher & Provider Gateway]
        Sanitizer[State-Machine JSON Sanitizer]
        
        Gin --> AICtrl
        Gin --> WSHub
        AICtrl --> Sanitizer
        Gin --> GORM
    end

    subgraph DataTier [Persistence & External AI Engines]
        direction LR
        MySQL[(MySQL Enterprise DB)]
        Groq[Groq Whisper STT API]
        NVIDIA[NVIDIA NIM / OpenAI API]
        Gemini[Gemini Vision OCR API]
    end

    Client <-->|HTTP / Websocket WS| APIServer
    Client <-->|Web Routes| WebServer
    GORM <--> DataTier
    AICtrl <-->|REST multipart/json| Groq
    AICtrl <-->|Chat Completions| NVIDIA
    AICtrl <-->|Multimodal Input| Gemini
```

---

## 2. Key Features & Technological Innovations

### 2.1 AI Dosen Persona Guardrails (Anti-Hallucination Engine)
The system injects specialized academic prompts (`personaDosenPrompt`) into LLM inference pipelines to ensure that AI recommendations remain 100% truthful to the supervisor's actual feedback:
- **Zero New Ideas Policy**: AI is strictly prohibited from inventing new research topics, methodologies, or corrections not explicitly uttered by the lecturer in guidance recordings or annotated documents.
- **Academic Classification Structure**: Revision items are automatically categorized into two standard academic domains:
  - **HOC (Higher Order Concerns / Major)**: Structural research issues (e.g., hypothesis validity, research model alignment, methodologies, data analysis).
  - **LOC (Lower Order Concerns / Minor)**: Technical writing mechanics (e.g., typographical errors, citation formatting APA/IEEE, grammatical structure, layout margins).

### 2.2 Groq Whisper Audio Slicing Engine (STT)
When students upload audio recordings of guidance sessions (`.mp3`, `.wav`):
- **Dynamic File Chunking**: If the uploaded audio exceeds **20 MB** (below Groq's 25MB safety threshold), the backend (`ai_controller.go` -> `transcribeAudio`) automatically splits the binary file into ordered byte slices, transmits them concurrently to Groq Whisper (`whisper-large-v3`), and seamlessly stitches the transcripts together.
- **Graceful Fallback**: If `GROQ_API_KEY` is not set in environment variables, the system prevents application crashes and bypasses transcription with an informative log while keeping document inspection features operational.

### 2.3 Multimodal Vision OCR & Track Changes Parsing
- **Gemini Vision OCR (`gemini-2.0-flash`)**: Extracts handwritten notes, red-pen annotations, and marginal remarks from uploaded photos of physical papers.
- **DOCX Revision Parsing**: Native parsing of Word `.docx` track changes to convert inline editor edits directly into actionable student tasks.

### 2.4 Resilient State-Machine JSON Sanitizer
LLMs frequently return JSON blocks containing unescaped raw newlines (`\n`, `\r`, `\t`) inside string literals, breaking default JSON parsers (`json.Unmarshal` throwing `invalid character '\n' in string literal`).
TierLog implements a custom **state-machine JSON sanitizer** (`sanitizeJSON`) in Go that parses string tokens char-by-char, escaping raw control characters inside quotes while preserving valid structural formatting.

### 2.5 Provider-Agnostic LLM Gateway & Dual Key Architecture
TierLog supports dual-purpose API keys:
1. **System-wide STT Key**: Global `GROQ_API_KEY` configured by server administrators for audio transcription.
2. **Per-User Encrypted LLM Keys**: Stored in the database for each user (`nvidia_key`, `openai_key`, `gemini_key`), allowing individual preference for preferred inference providers.
3. **Capability Filtering**: Intelligent endpoint capability discovery filtering models by multimodal capabilities (e.g., separating text-only LLMs from vision-capable models).

### 2.6 Realtime WebSocket Hub & Event Broadcasting
Built-in WebSocket hub (`ws://localhost:8080/ws`) enables live room broadcasting. Whenever a student uploads a revision or a lecturer updates consultation status, connected clients receive instant state updates without requiring page reloads.

### 2.7 Widescreen Desktop Workspace Engine (1600px)
To maximize productivity on modern widescreen monitors, the frontend features an expanded `1600px` (`size="xl"`) layout container on consultation pages, lecturer dashboards, and archives, supporting split-pane document reviewing.

---

## 3. Project Directory Structure

```
Tierlog_EXPO/
├── PopularProgramingFinalProject/         # Main Project Package
│   ├── main.go                            # Go Application Entry Point & Route Declarations
│   ├── Dockerfile                         # Go Backend Container Build Spec
│   ├── docker-compose.yml                 # Multi-container Orchestration (Go + Laravel + MySQL)
│   ├── go.mod / go.sum                    # Go Dependency Manifests
│   ├── struct_go.sql                      # Database Schema Initialization Dump
│   │
│   ├── controller/                        # Gin Request Controllers
│   │   ├── ai_controller.go               # AI Inference Gateway, Whisper Chunking, JSON Sanitizer
│   │   ├── app_controller.go              # Consultation CRUD & File Management
│   │   ├── annotation_ctrl.go             # Annotation & Image Upload Handlers
│   │   └── user_controller.go             # Authentication & API Key Management
│   │
│   ├── models/                            # GORM Database Struct Models
│   │   ├── user.go                        # User Account & Encrypted Credentials
│   │   ├── consultation.go                # Consultation Logs & Status Trackers
│   │   └── annotation.go                  # Revision Annotations & Tasks
│   │
│   ├── middleware/                        # JWT & Security Middlewares
│   ├── realtime/                          # WebSocket Hub & Room Broadcaster
│   ├── utils/                             # Crypto & Encryption Helpers
│   │
│   └── tierlog_web/                       # Frontend Web Application (Laravel 11 + Inertia)
│       ├── artisan                        # Laravel CLI Tool
│       ├── package.json                   # Vite, Vue 3, Pinia, Tailwind Dependencies
│       ├── Dockerfile                     # Frontend Web Container Build Spec
│       ├── app/                           # Laravel Controllers & Middleware
│       ├── config/                        # Framework Configurations
│       ├── database/                      # SQLite / Migrations
│       ├── resources/
│       │   ├── js/                        # Vue 3 Components & TypeScript Modules
│       │   │   ├── Pages/                 # Inertia Page Views (Consultations, Dashboard, Archive)
│       │   │   ├── components/            # Reusable Workspace UI Components
│       │   │   └── types/                 # TypeScript Interface Definitions
│       │   └── css/                       # Tailwind CSS v4 Stylesheets
│       └── routes/                        # Web & Auth Routes
```

---

## 4. Database Architecture & Data Models

TierLog utilizes MySQL with GORM Object-Relational Mapping. Below is the relational core entity model:

### Key Tables

#### 1. `users`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UINT | Primary Key, Auto Increment | Unique user identifier |
| `name` | VARCHAR(255) | NOT NULL | Full name of student/lecturer |
| `email` | VARCHAR(255) | Unique, NOT NULL | Account login email |
| `password` | VARCHAR(255) | NOT NULL | Bcrypt hashed password |
| `role` | VARCHAR(50) | NOT NULL | Account role (`student`, `lecturer`, `admin`) |
| `preferred_model`| VARCHAR(100) | Default: `nvidia:llama-3.2` | Preferred LLM provider:model string |
| `groq_key` | VARCHAR(255) | Encrypted | Per-user or override Groq API key |
| `nvidia_key` | VARCHAR(255) | Encrypted | Per-user NVIDIA NIM API key |

#### 2. `consultations`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UINT | Primary Key, Auto Increment | Consultation session ID |
| `student_id` | UINT | Foreign Key (`users.id`) | Student submission owner |
| `lecturer_id` | UINT | Foreign Key (`users.id`) | Assigned supervisor |
| `title` | VARCHAR(255) | NOT NULL | Thesis chapter or submission title |
| `status` | VARCHAR(50) | Default: `pending` | Session state (`pending`, `reviewed`, `completed`) |
| `audio_path` | VARCHAR(500) | Optional | Storage path to uploaded guidance audio |
| `paper_path` | VARCHAR(500) | NOT NULL | Storage path to thesis manuscript (.docx) |
| `transcript` | LONGTEXT | Optional | Stitched output from Groq Whisper STT |

#### 3. `annotations`
| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| `id` | UINT | Primary Key, Auto Increment | Annotation item ID |
| `consultation_id`| UINT | Foreign Key (`consultations.id`)| Parent consultation session |
| `category` | VARCHAR(20) | Enum (`HOC`, `LOC`) | Academic severity classification |
| `feedback_text` | TEXT | NOT NULL | Revision item detail |
| `is_completed` | BOOLEAN | Default: `false` | Student completion checklist status |

---

## 5. API & WebSocket Contract Specifications

### REST API Endpoints (Go Gateway: Port 8080)

#### Authentication & Profile
- `POST /api/v1/register` — Register new user account.
- `POST /api/v1/login` — Authenticate and obtain JWT bearer token.
- `GET /api/v1/profile` — Fetch current authenticated profile & API key configurations.
- `PUT /api/v1/profile/keys` — Update encrypted provider API keys (`groq_key`, `nvidia_key`, etc.).

#### Consultation Management
- `GET /api/v1/consultations` — List consultations (filtered by user role).
- `POST /api/v1/consultations` — Create consultation session (supports `multipart/form-form` with audio and document files).
- `GET /api/v1/consultations/:id` — Retrieve consultation details, transcript, and AI revision checklist.
- `DELETE /api/v1/consultations/:id` — Purge consultation session and clean up physical disk storage.

#### AI Processing & Models
- `GET /api/v1/ai/models` — Discover available models filtered by provider capability (`text` vs `vision`).
- `POST /api/v1/ai/analyze` — Trigger manual re-analysis of consultation documents.

### Realtime WebSocket Protocol
- **Endpoint**: `ws://localhost:8080/ws`
- **Handshake**: Connect with query token `ws://localhost:8080/ws?token=<JWT_TOKEN>`.
- **Event Payload Structure**:
  ```json
  {
    "event": "CONSULTATION_UPDATED",
    "room_id": "consultation_42",
    "data": {
      "id": 42,
      "status": "reviewed",
      "updated_at": "2026-06-28T15:43:00Z"
    }
  }
  ```

---

## 6. Installation & Environment Configuration

### Prerequisites
- **Go**: Version 1.22 or higher
- **PHP / Composer**: PHP 8.2+ and Composer 2.x
- **Node.js**: Version 18+ and npm / pnpm
- **MySQL**: Version 8.0+

### Environment Files Setup

1. **Go Backend Environment (`PopularProgramingFinalProject/.env`)**:
   ```env
   PORT=8080
   DB_HOST=127.0.0.1
   DB_PORT=3306
   DB_USER=root
   DB_PASSWORD=your_password
   DB_NAME=tierlog_db
   JWT_SECRET=your_super_secret_jwt_key_enterprise
   
   # Optional Fallback Keys
   GROQ_API_KEY=gsk_your_groq_api_key_here
   NVIDIA_API_KEY=nvapi_your_nvidia_api_key_here
   ```

2. **Laravel Frontend Environment (`PopularProgramingFinalProject/tierlog_web/.env`)**:
   ```env
   APP_NAME=TierLog
   APP_ENV=local
   APP_KEY=base64:generated_app_key_here
   APP_URL=http://localhost:8000
   
   VITE_GO_BACKEND_URL=http://localhost:8080
   VITE_WS_BACKEND_URL=ws://localhost:8080/ws
   ```

### Running Locally

1. **Start MySQL Database**: Ensure MySQL is running and create database `tierlog_db`. Import schema from `PopularProgramingFinalProject/struct_go.sql`.
2. **Start Go Backend**:
   ```bash
   cd PopularProgramingFinalProject
   go run main.go
   ```
3. **Start Laravel Frontend**:
   ```bash
   cd PopularProgramingFinalProject/tierlog_web
   composer install
   npm install
   npm run dev
   php artisan serve --port=8000
   ```

---

## 7. Production Deployment & Containerization

TierLog includes enterprise Docker configurations for unified containerized orchestration.

### Deploying with Docker Compose

To deploy the entire multi-service stack (Go Backend + Laravel Web + MySQL DB):

```bash
cd PopularProgramingFinalProject
docker-compose up -d --build
```

The services will spin up automatically on the specified ports:
- **Web Interface**: `http://localhost:8000`
- **Go API Gateway**: `http://localhost:8080`
- **MySQL Database**: `localhost:3306`

---

## 8. Enterprise Security & Reliability Standard

- **Encrypted Secrets at Rest**: User API keys are symmetrically encrypted before storage in MySQL using AES-GCM primitives (`utils/crypto.go`).
- **Input Sanitization & MIME Guard**: Uploaded files are verified via magic-byte checking to prevent arbitrary executable execution.
- **Circuit-Breaker Pattern for AI APIs**: External calls to Groq and NVIDIA NIM feature configurable timeouts and retry policies to prevent service hangs during API outages.

---

*© 2026 TierLog Platform Team. All Rights Reserved. Enterprise Academic Systems.*
