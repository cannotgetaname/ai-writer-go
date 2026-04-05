# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build Go backend
go build -o ai-writer .

# Run directly without building
go run . server

# Run web server (default port 8081)
./ai-writer server
./ai-writer server -p 9090 -H 0.0.0.0  # custom port/host
./ai-writer server -d                   # daemon mode

# Build Vue frontend
cd web && npm install && npm run build

# Develop frontend with hot reload
cd web && npm run dev

# Run CLI commands
./ai-writer book list
./ai-writer -b <book_name> chapter list
./ai-writer -b <book_name> write 1 --stream
```

**Note**: No test files exist in this codebase yet.

## Architecture Overview

**Backend (Go)**:
- `cmd/` - CLI commands using Cobra framework. `root.go` has global flags (`-b` for book, `-c` for config)
- `internal/api/` - Gin REST API with handlers in `handler/handler.go`, router in `router.go`
- `internal/store/json_store.go` - JSON file storage layer, all data persistence goes through here
- `internal/service/` - Business logic layer (writer, reviewer, architect, context, graph, sync, toolbox, analysis)
- `internal/llm/` - LLM client abstraction supporting DeepSeek (default)/OpenAI/Ollama
- `internal/model/` - Data structures (Book, Chapter, Character, Item, Location, Foreshadow, CausalChain, NarrativeThread)
- `internal/engine/` - Advanced narrative features (causal_chain, narrative_thread, emotional_arc, consistency, info_boundary)
- `internal/config/config.go` - Configuration with defaults and prompt templates, loaded from `config.yaml`

**Frontend (Vue 3)**:
- `web/src/api/index.js` - Axios API client, all endpoints organized by domain (bookApi, chapterApi, aiApi, etc.)
- `web/src/views/` - Vue components for each page (WritingView, BatchView, SyncView, ToolboxView, etc.)
- `web/src/router.js` - Vue Router configuration
- Uses Element Plus UI, ECharts for graphs, Pinia for state management

## Data Storage Structure

All book data stored in JSON files under `data/projects/{book_name}/`:
```
data/projects/{book_name}/
├── metadata.json        # Book metadata
├── structure.json       # Chapters (array)
├── volumes.json         # Volumes
├── characters.json      # Characters
├── items.json           # Items
├── locations.json       # Locations
├── worldview.json       # World settings
├── foreshadows.json     # Foreshadowing tracking
├── causal_chains.json   # Causal events
├── threads.json         # Narrative threads
└── chapters/
    ├── 1.json           # Chapter 1 paragraphs (ChapterParagraphs struct)
    ├── 2.json           # Chapter 2 paragraphs
    └── ...
```

**Important**: Chapter content uses paragraph-based storage (`ChapterParagraphs`), not plain text. Each paragraph has unique ID for editing/tracking.

## Key Patterns

### Store Layer (`internal/store/json_store.go`)
All data operations go through JSONStore methods:
- `LoadX/SaveX` pattern for each entity type (books, chapters, characters, items, locations, worldview, foreshadows, causal_chains, threads)
- `LoadChapterParagraphs/SaveChapterParagraphs` for chapter content
- Uses mutex (`sync.RWMutex`) for concurrent access protection
- Billing tracked via `billing_store.go` with `RecordUsage/GetStats` methods

### Service Layer (`internal/service/`)
Business logic separated from handlers:
- `WriterService` - AI content generation with context building, supports both sync and stream output
- `ContextService` - Builds writing context from worldview, characters, previous chapters, and summaries
- `GraphService` - Builds knowledge graph from entities for ECharts visualization
- `SyncService` - Extracts state changes from chapter content, manages pending updates
- `ToolboxService` - Naming, character, conflict, scene generation tools

### Handler Layer (`internal/api/handler/handler.go`)
- `InitStore()` and `InitConfig()` called from router to set global instances
- Returns JSON responses, handles errors with appropriate HTTP codes
- For AI endpoints, creates LLM client via `getLLMClient()` helper using config's task-based model routing

### LLM Integration (`internal/llm/`)
- `Client` interface with `Call/CallWithSystem/Stream/StreamWithSystem` methods
- Task-based routing: different models/temperatures per task type (writer: 1.5, architect: 1.0, reviewer: 0.5)
- Providers: DeepSeek (default), OpenAI, Ollama - all use OpenAI-compatible API format
- Stream output via `StreamChunk` channel with `Content`, `Done`, `Error` fields

### State Sync Pattern
After AI generates chapters, `SyncService` extracts implicit state changes (new characters, item ownership changes, location discoveries). Users review pending updates via `/api/sync/pending` then apply or reject them.

## API Routes Summary

See `internal/api/router.go` for full routing. Key endpoints:
- `/api/books` - CRUD for books
- `/api/books/:id/chapters` - Chapter management
- `/api/books/:id/settings/*` - Worldview, characters, items, locations
- `/api/books/:id/graph/echarts` - Knowledge graph data for ECharts
- `/api/books/:id/foreshadows` - Foreshadowing tracking (create, resolve, warnings)
- `/api/books/:id/causal-chain` - Causal event chain management
- `/api/ai/*` - AI generation, review, audit, rewrite
- `/api/batch/*` - Batch chapter generation with status tracking
- `/api/sync/*` - Extract state changes from chapters, apply/reject pending updates
- `/api/toolbox/*` - Writing tools (naming, conflict, scene, goldfinger, title, synopsis, twist, dialogue)
- `/api/architect/*` - Plot architecture generation and fission
- `/api/analysis/*` - External book parsing and analysis (拆书分析)
- `/api/system/*` - Config, prompts, billing stats, writing goals

## Configuration

`config.yaml` contains:
- `llm.provider`: openai/deepseek/ollama
- `llm.api_key`: API key (can also use environment variable)
- `llm.models`: Task-to-model mapping (writer, architect, reviewer, editor, auditor, timekeeper, summary)
- `llm.temperatures`: Task-specific temperatures
- `server.port`: Web server port

**Important**: No `config.example.yaml` exists. Default configuration including all prompt templates is defined in `internal/config/config.go` under `defaultConfig`. The prompts include specialized system prompts for:
- `WriterSystem` - Content generation with "show don't tell" principles
- `ArchitectSystem` - Plot planning with JSON output
- `ReviewerSystem` - Critical review for consistency and pacing
- `AuditorSystem` - State change extraction (characters, items, locations)
- `TimekeeperSystem` - Timeline tracking

**Billing**: Pricing rates per model are tracked in `defaultConfig.Pricing` (DeepSeek, GPT-4, Claude). Token costs are calculated via `CalculateCost()` and stored in `internal/store/billing_store.go`.