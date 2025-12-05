# Karbos

**Carbon-Aware Job Scheduling Platform**

Karbos is an intelligent job scheduling system that optimizes workload execution based on real-time grid carbon intensity, reducing environmental impact while meeting SLA requirements.

## 🌍 Mission

Schedule batch jobs during periods of low carbon intensity to reduce the environmental footprint of cloud computing without compromising performance or deadlines.

## 🏗️ Architecture

```
┌─────────────────┐
│  Dashboard      │  Next.js 15 + React 19 + TypeScript
│  (Port 3000)    │  Framer Motion animations
└────────┬────────┘
         │ HTTP
         ▼
┌─────────────────┐
│  API Gateway    │  Go + Fiber
│  (Port 8080)    │  Job submission & tracking
└────────┬────────┘
         │
    ┌────┴────┐
    ▼         ▼
┌─────────┐ ┌─────────┐
│Postgres │ │  Redis  │
│Database │ │  Queue  │
└─────────┘ └─────────┘
```

## 📦 Project Structure

```
Karbos/
├── client/                 # Frontend Dashboard (Next.js)
│   ├── app/               # Next.js 15 App Router
│   ├── components/        # React components with animations
│   │   ├── Navigation.tsx
│   │   └── tabs/
│   │       ├── Overview.tsx        # KPIs & Eco-Curve
│   │       ├── Workloads.tsx       # Job queue
│   │       ├── GridIntelligence.tsx # Carbon forecasts
│   │       ├── Infrastructure.tsx   # Worker nodes
│   │       └── Playground.tsx      # Job submission
│   ├── lib/               # Utilities
│   └── types/             # TypeScript definitions
│
├── server/                # Backend API (Go)
│   ├── cmd/api/           # Server entry point
│   ├── internal/
│   │   ├── config/        # Configuration
│   │   ├── database/      # PostgreSQL operations
│   │   ├── handlers/      # HTTP handlers
│   │   ├── models/        # Data models
│   │   └── queue/         # Redis queue
│   └── database/
│       └── schema.sql     # Database schema
│
├── docs/                  # Documentation
└── audit-logs/            # Project audit trail
```

## 🚀 Phase 1: Infrastructure Skeleton ✅

**Status:** Complete

### What's Built

#### Client (Dashboard)
- ✅ 5 interactive tabs with professional animations
- ✅ Real-time KPI cards with metrics
- ✅ 24-hour carbon intensity Eco-Curve chart
- ✅ Job queue table with status tracking
- ✅ Regional carbon intensity map
- ✅ Worker node infrastructure monitoring
- ✅ Job submission playground with simulation
- ✅ Framer Motion animations throughout

#### Server (API)
- ✅ PostgreSQL database schema (Jobs, Logs, Carbon Cache)
- ✅ Redis message broker (Immediate + Delayed queues)
- ✅ HTTP API Gateway with Fiber
- ✅ Job submission endpoint with validation
- ✅ Health check and monitoring endpoints
- ✅ UUID-based job identification
- ✅ Enum-based status tracking

### Tech Stack

**Frontend:**
- Next.js 15 (App Router)
- React 19
- TypeScript (strict mode)
- Tailwind CSS
- Framer Motion

**Backend:**
- Go 1.21+
- Fiber v2 (HTTP framework)
- PostgreSQL (Supabase compatible)
- Redis (message broker)
- google/uuid

## 🎯 Getting Started

### Prerequisites

- Node.js 18+
- Go 1.21+
- PostgreSQL (or Supabase account)
- Redis

### Quick Start

#### 1. Setup Client (Dashboard)

```bash
cd client
npm install
npm run dev
# Visit http://localhost:3000
```

#### 2. Setup Server (API)

```bash
cd server

# Copy environment template
cp .env.example .env

# Edit .env with your credentials
# DATABASE_URL=postgresql://...
# REDIS_HOST=localhost

# Setup database
psql -d karbos -f database/schema.sql

# Start Redis
docker run -d -p 6379:6379 redis:alpine

# Run server
go run cmd/api/main.go
# Server runs on http://localhost:8080
```

#### 3. Test the System

```bash
# Submit a job
curl -X POST http://localhost:8080/api/submit \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "demo-user",
    "docker_image": "python:3.11",
    "deadline": "2025-12-06T18:00:00Z"
  }'

# Check health
curl http://localhost:8080/health
```

## 📚 Documentation

- **[Client README](./client/README.md)** - Dashboard setup and features
- **[Server README](./server/README.md)** - API documentation
- **[Setup Guide](./server/SETUP_GUIDE.md)** - Detailed installation
- **[Phase 1 Summary](./server/PHASE1_SUMMARY.md)** - What's been built
- **[Database Schema](./server/database/schema.sql)** - SQL schema

## 🎨 Features

### Dashboard
- **Overview Tab**: CO₂ savings metrics, Eco-Curve forecast, recent activity
- **Workloads Tab**: Job queue with status, drawer details, execution logs
- **Grid Intelligence Tab**: Regional carbon intensity map, generation mix
- **Infrastructure Tab**: Worker nodes, queue health, system metrics
- **Playground Tab**: Interactive job submission with simulation

### API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/submit` | Submit new job |
| GET | `/api/jobs/:id` | Get job details |
| GET | `/api/users/:userId/jobs` | List user jobs |
| GET | `/health` | Service health check |
| GET | `/ready` | Readiness probe |

## 🧪 Testing

### Frontend
```bash
cd client
npm run build  # Verify build
npm run dev    # Development server
```

### Backend
```bash
cd server
go test ./...              # Run tests
go build cmd/api/main.go   # Build binary
```

### API Testing
- Import `server/postman_collection.json` into Postman
- Or use cURL commands from documentation

## 📊 Database Schema

### Jobs Table
```sql
- id (UUID, Primary Key)
- user_id (VARCHAR)
- docker_image (VARCHAR)
- status (ENUM: PENDING, DELAYED, RUNNING, COMPLETED, FAILED)
- scheduled_time (TIMESTAMP)
- deadline (TIMESTAMP)
- created_at, started_at, completed_at
```

### Redis Queue Structure
- **Immediate Queue**: `karbos:queue:immediate` (List/FIFO)
- **Delayed Set**: `karbos:queue:delayed` (Sorted Set by timestamp)

## 🎯 Next Phase (Phase 2)

Coming next:
- [ ] Carbon intensity API integration (WattTime, Electricity Maps)
- [ ] Scheduling optimization algorithm
- [ ] Worker execution engine
- [ ] Delayed job promotion logic
- [ ] Real-time carbon calculations
- [ ] WebSocket updates for live dashboard

## 🛠️ Development

### Client Commands
```bash
npm run dev          # Start dev server
npm run build        # Production build
npm run lint         # Run ESLint
```

### Server Commands
```bash
go run cmd/api/main.go    # Run server
go test ./...             # Run tests
go build -o bin/karbos    # Build binary
```

## 🎨 Design System

**Color Palette (Karbos Brand):**
- Navy: `#1A1B41`
- Indigo: `#292E6F`
- Blue-Purple: `#525FB0`
- Lavender: `#A6B1E1`
- Light Blue: `#D6DFFF`

## 📝 License

MIT

## 👥 Contributing

This is a demonstration project showcasing:
- Modern full-stack architecture
- Carbon-aware computing concepts
- Go microservices
- React 19 + Next.js 15
- Professional UI/UX with animations

## 🙏 Acknowledgments

- Built as a learning project for sustainable computing
- Inspired by carbon-aware computing initiatives
- Dashboard design follows industry best practices

---

**Status**: Phase 1 Complete ✅  
**Version**: 1.0.0  
**Last Updated**: December 5, 2025

For detailed setup instructions, see [SETUP_GUIDE.md](./server/SETUP_GUIDE.md)
