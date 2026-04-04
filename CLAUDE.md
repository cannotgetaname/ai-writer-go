# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build and Run Commands

```bash
# Build Go backend
go build -o ai-writer .

# Run web server (default port 8081)
./ai-writer server
./ai-writer server -p 9090 -H 0.0.0.0  # custom port/host
./ai-writer server -d                   # daemon mode

# Build Vue frontend
cd web && npm run build

# Develop frontend with hot reload
cd web && npm run dev

# Run CLI commands
./ai-writer book list
./ai-writer -b <book_name> chapter list
./ai-writer -b <book_name> write 1 --stream
```

## Architecture Overview

**Backend (Go)**:
- `cmd/` - CLI commands using Cobra framework. `root.go` has global flags (`-b` for book, `-c` for config)
- `internal/api/` - Gin REST API with handlers in `handler/handler.go`
- `internal/store/json_store.go` - JSON file storage layer, all data persistence goes through here
- `internal/service/` - Business logic layer (writer, reviewer, architect, context, graph)
- `internal/llm/` - LLM client abstraction supporting OpenAI/DeepSeek/Ollama
- `internal/model/` - Data structures (Book, Chapter, Character, Item, Location, etc.)
- `internal/engine/` - Advanced features (causal chains, narrative threads, emotional arcs)
- `internal/config/config.go` - Configuration with defaults, loaded from `config.yaml`

**Frontend (Vue 3)**:
- `web/src/api/index.js` - Axios API client, all endpoints defined here
- `web/src/views/` - Vue components for each page
- Uses Element Plus UI, ECharts for graphs, Pinia for state

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
- `LoadX/SaveX` pattern for each entity type
- `LoadChapterParagraphs/SaveChapterParagraphs` for chapter content
- Uses mutex for concurrent access protection

### Service Layer (`internal/service/`)
Business logic separated from handlers:
- `WriterService` - AI content generation with context building
- `ContextService` - Builds writing context from worldview, characters, previous chapters
- `GraphService` - Builds knowledge graph from entities for ECharts

### Handler Layer (`internal/api/handler/handler.go`)
- `InitStore()` called from router to set global store instance
- Returns JSON responses, handles errors with appropriate HTTP codes
- For AI endpoints, creates LLM client via `getLLMClient()` helper

### LLM Integration (`internal/llm/`)
- `Client` interface with `Call/CallWithSystem/Stream/StreamWithSystem` methods
- Task-based routing: different models/temperatures per task (writer, reviewer, architect, etc.)
- Config in `config.yaml` under `llm.models` and `llm.temperatures`

## API Routes Summary

See `internal/api/router.go` for full routing. Key endpoints:
- `/api/books` - CRUD for books
- `/api/books/:id/chapters` - Chapter management
- `/api/books/:id/settings/*` - Worldview, characters, items, locations
- `/api/books/:id/graph/echarts` - Knowledge graph data for ECharts
- `/api/ai/*` - AI generation, review, audit, rewrite
- `/api/toolbox/*` - Writing tools (naming, conflict, scene, etc.)
- `/api/system/*` - Config, prompts, billing stats

## Configuration

`config.yaml` contains:
- `llm.provider`: openai/deepseek/ollama
- `llm.api_key`: API key (can also use environment variable)
- `llm.models`: Task-to-model mapping (writer, architect, reviewer, etc.)
- `llm.temperatures`: Task-specific temperatures
- `server.port`: Web server port

Default prompts are defined in `internal/config/config.go` under `defaultConfig.Prompts`.