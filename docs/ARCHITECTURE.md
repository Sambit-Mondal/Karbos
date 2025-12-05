# Karbos System Architecture

## High-Level Overview

```
┌────────────────────────────────────────────────────────┐
│                    User Interface                       │
│              Next.js Dashboard (Port 3000)              │
│  ┌──────────┬──────────┬──────────┬──────────────┐    │
│  │ Overview │Workloads │   Grid   │Infrastructure│    │
│  │          │          │Intelligence│   Playground │    │
│  └──────────┴──────────┴──────────┴──────────────┘    │
└────────────────────┬───────────────────────────────────┘
                     │ HTTP/REST API
                     │ JSON
                     ▼
┌────────────────────────────────────────────────────────┐
│                  API Gateway Layer                      │
│              Go Fiber Server (Port 8080)                │
│  ┌──────────────────────────────────────────────────┐  │
│  │  Routes: /api/submit, /api/jobs, /health        │  │
│  │  Middleware: CORS, Logging, Recovery, RequestID │  │
│  └──────────────────────────────────────────────────┘  │
└──────────┬─────────────────────────┬───────────────────┘
           │                         │
           │                         │
           ▼                         ▼
┌─────────────────────┐   ┌─────────────────────┐
│  PostgreSQL/Supabase│   │     Redis Cache      │
│  (Database Layer)   │   │  (Message Broker)    │
│                     │   │                      │
│ ┌─────────────────┐ │   │ ┌────────────────┐ │
│ │ Jobs Table      │ │   │ │ Immediate Queue│ │
│ │ - ID (UUID)     │ │   │ │  (List/FIFO)   │ │
│ │ - Status (Enum) │ │   │ └────────────────┘ │
│ │ - Deadline      │ │   │                    │
│ └─────────────────┘ │   │ ┌────────────────┐ │
│                     │   │ │  Delayed Set   │ │
│ ┌─────────────────┐ │   │ │  (ZSet/Score)  │ │
│ │ Execution Logs  │ │   │ └────────────────┘ │
│ └─────────────────┘ │   │                    │
│                     │   │  Connection Pool   │
│ ┌─────────────────┐ │   │  Health Checks     │
│ │ Carbon Cache    │ │   └────────────────────┘
│ └─────────────────┘ │
└─────────────────────┘
```

## Data Flow - Job Submission

```
┌─────────┐
│  User   │
│ Browser │
└────┬────┘
     │ 1. User fills job submission form
     │
     ▼
┌─────────────────┐
│ Playground Tab  │
│ (React Form)    │
└────┬────────────┘
     │ 2. POST /api/submit
     │    {user_id, docker_image, deadline}
     │
     ▼
┌─────────────────────┐
│   Job Handler       │
│ (Go/Fiber)          │
│                     │
│ 1. Validate request │
│ 2. Check deadline   │
│ 3. Generate UUID    │
└────┬────────────────┘
     │
     │ 3. INSERT INTO jobs
     ▼
┌─────────────────────┐
│  Job Repository     │
│                     │
│ CreateJob()         │
│ - Save to Postgres  │
│ - Set status=PENDING│
└────┬────────────────┘
     │
     │ 4. Enqueue job
     ▼
┌─────────────────────┐
│   Redis Queue       │
│                     │
│ RPUSH immediate_queue│
│ {job_id, image...} │
└────┬────────────────┘
     │
     │ 5. Return response
     ▼
┌─────────────────────┐
│  HTTP Response      │
│                     │
│ {job_id, status,    │
│  created_at}        │
└─────────────────────┘
```

## Component Architecture

### Frontend (Client)

```
client/
│
├── app/
│   ├── layout.tsx ────────────► Root layout with metadata
│   ├── globals.css ───────────► Tailwind & custom styles
│   └── page.tsx ──────────────► Main page with tab routing
│
├── components/
│   ├── Navigation.tsx ────────► Top nav bar with tabs
│   │                             • Tab switching
│   │                             • Region selector
│   │                             • System status
│   │
│   └── tabs/
│       ├── Overview.tsx ──────► Dashboard home
│       │                         • KPI cards
│       │                         • Eco-Curve chart
│       │                         • Recent activity
│       │
│       ├── Workloads.tsx ─────► Job management
│       │                         • Job queue table
│       │                         • Details drawer
│       │                         • Status tracking
│       │
│       ├── GridIntelligence.tsx ► Carbon data
│       │                         • Regional map
│       │                         • Intensity forecasts
│       │                         • Generation mix
│       │
│       ├── Infrastructure.tsx ► System health
│       │                         • Worker nodes
│       │                         • Queue metrics
│       │                         • Resource usage
│       │
│       └── Playground.tsx ────► Job submission
│                                 • Form inputs
│                                 • Simulation mode
│                                 • Results display
│
├── lib/
│   └── utils.ts ──────────────► Helper functions
│
└── types/
    └── index.ts ──────────────► TypeScript types
```

### Backend (Server)

```
server/
│
├── cmd/api/
│   └── main.go ───────────────► Server entry point
│                                • Initialize DB
│                                • Initialize Redis
│                                • Setup routes
│                                • Start server
│
├── internal/
│   │
│   ├── config/
│   │   └── config.go ─────────► Configuration loader
│   │                             • Env variables
│   │                             • Validation
│   │
│   ├── models/
│   │   └── models.go ─────────► Data structures
│   │                             • Job
│   │                             • ExecutionLog
│   │                             • CarbonCache
│   │                             • Request/Response
│   │
│   ├── database/
│   │   ├── database.go ───────► DB connection
│   │   │                         • Connection pool
│   │   │                         • Health check
│   │   │
│   │   └── job_repository.go ► Job operations
│   │                             • CreateJob
│   │                             • GetJobByID
│   │                             • UpdateJobStatus
│   │                             • GetJobsByStatus
│   │
│   ├── queue/
│   │   └── redis_queue.go ────► Queue operations
│   │                             • EnqueueImmediate
│   │                             • EnqueueDelayed
│   │                             • DequeueImmediate
│   │                             • GetDueJobs
│   │
│   └── handlers/
│       ├── job_handler.go ────► Job endpoints
│       │                         • SubmitJob
│       │                         • GetJob
│       │                         • GetUserJobs
│       │
│       └── health_handler.go ► Health checks
│                                 • HealthCheck
│                                 • ReadyCheck
│
└── database/
    └── schema.sql ────────────► Database schema
                                  • Tables
                                  • Indexes
                                  • Triggers
```

## Request/Response Flow

### POST /api/submit

```
Request:
┌──────────────────────────────┐
│ POST /api/submit             │
│ Content-Type: application/json│
│                              │
│ {                            │
│   "user_id": "user-123",     │
│   "docker_image": "py:3.11", │
│   "deadline": "2025-12-06...",│
│   "estimated_duration": 300  │
│ }                            │
└──────────────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Validation Layer             │
│ • Check required fields      │
│ • Parse deadline (RFC3339)   │
│ • Verify deadline > now      │
└──────────────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Business Logic               │
│ • Generate UUID              │
│ • Set status = PENDING       │
│ • Create job object          │
└──────────────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Database Persistence         │
│ INSERT INTO jobs (...)       │
│ RETURNING id, created_at     │
└──────────────────────────────┘
         │
         ▼
┌──────────────────────────────┐
│ Queue Operations             │
│ RPUSH karbos:queue:immediate │
│ JSON.stringify(queueItem)    │
└──────────────────────────────┘
         │
         ▼
Response:
┌──────────────────────────────┐
│ HTTP 201 Created             │
│                              │
│ {                            │
│   "job_id": "550e8400-...",  │
│   "status": "PENDING",       │
│   "created_at": "2025-...",  │
│   "message": "Job submitted" │
│ }                            │
└──────────────────────────────┘
```

## Database Schema Relationships

```
┌─────────────────────────────────────┐
│            jobs                      │
├─────────────────────────────────────┤
│ id (UUID) PRIMARY KEY               │◄──┐
│ user_id (VARCHAR)                   │   │
│ docker_image (VARCHAR)              │   │
│ status (job_status ENUM)            │   │
│ scheduled_time (TIMESTAMPTZ)        │   │
│ deadline (TIMESTAMPTZ)              │   │
│ created_at (TIMESTAMPTZ)            │   │
└─────────────────────────────────────┘   │
                                          │
                                          │ Foreign Key
                                          │
┌─────────────────────────────────────┐   │
│      execution_logs                  │   │
├─────────────────────────────────────┤   │
│ id (UUID) PRIMARY KEY               │   │
│ job_id (UUID) ──────────────────────┼───┘
│ output (TEXT)                       │
│ exit_code (INTEGER)                 │
│ duration (INTEGER)                  │
│ started_at (TIMESTAMPTZ)            │
│ completed_at (TIMESTAMPTZ)          │
└─────────────────────────────────────┘

┌─────────────────────────────────────┐
│        carbon_cache                  │
├─────────────────────────────────────┤
│ id (UUID) PRIMARY KEY               │
│ region (VARCHAR)                    │
│ timestamp (TIMESTAMPTZ)             │
│ intensity_value (DECIMAL)           │
│ forecast_window (INTEGER)           │
│ source (VARCHAR)                    │
└─────────────────────────────────────┘
```

## Redis Queue Structure

```
Immediate Queue (List):
┌────────────────────────────────────┐
│ karbos:queue:immediate             │
│                                    │
│ [newest] ←                → [oldest]│
│    │                          │     │
│    ▼                          ▼     │
│  job3    job2    job1              │
│                                    │
│ RPUSH (enqueue) →                  │
│                 ← LPOP (dequeue)   │
└────────────────────────────────────┘

Delayed Set (Sorted Set):
┌────────────────────────────────────┐
│ karbos:queue:delayed               │
│                                    │
│ Score (Unix)    Member (JSON)      │
│ ─────────────   ────────────────   │
│ 1733443200  →  {"job_id": "abc"}  │
│ 1733446800  →  {"job_id": "def"}  │
│ 1733450400  →  {"job_id": "ghi"}  │
│                                    │
│ ZADD score member                  │
│ ZRANGEBYSCORE -inf now             │
└────────────────────────────────────┘
```

## Technology Stack

```
┌──────────────────────────────────────────┐
│           Frontend Layer                  │
│  • Next.js 15 (App Router)               │
│  • React 19 (Server/Client Components)   │
│  • TypeScript (Strict Mode)              │
│  • Tailwind CSS (Utility-First)          │
│  • Framer Motion (Animations)            │
└──────────────────────────────────────────┘
                    ▼
┌──────────────────────────────────────────┐
│            API Layer                      │
│  • Go 1.21+ (Compiled Language)          │
│  • Fiber v2 (Web Framework)              │
│  • UUID (google/uuid)                    │
│  • Godotenv (Config)                     │
└──────────────────────────────────────────┘
                    ▼
┌────────────────┬─────────────────────────┐
│ Database Layer │  Message Broker Layer   │
│  • PostgreSQL  │  • Redis 7+             │
│  • lib/pq      │  • go-redis/v9          │
│  • Connection  │  • List/ZSet            │
│    Pool        │  • Pub/Sub ready        │
└────────────────┴─────────────────────────┘
```

## Deployment Architecture (Future)

```
                  ┌──────────────┐
                  │   Vercel     │
                  │  (Next.js)   │
                  └──────┬───────┘
                         │
              ┌──────────┴──────────┐
              │                     │
              ▼                     ▼
       ┌─────────────┐      ┌─────────────┐
       │   Fly.io    │      │   Render    │
       │ (Go Server) │      │ (Go Server) │
       └──────┬──────┘      └──────┬──────┘
              │                     │
    ┌─────────┴─────────────────────┴─────┐
    │                                      │
    ▼                                      ▼
┌─────────┐                          ┌─────────┐
│Supabase │                          │ Upstash │
│(Postgres)│                         │ (Redis) │
└─────────┘                          └─────────┘
```

---

**Last Updated**: December 5, 2025  
**Status**: Phase 1 Complete
