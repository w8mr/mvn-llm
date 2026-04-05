# Maven Tool (Node.js/TypeScript)

A modular CLI and OpenCode plugin for efficient, agent-friendly Maven tasks — build, test (with Surefire XML parsing), install, and dependency tree — with intelligent output summarization for agentic workflows.

## Features
- Supports: build, test (parses Surefire XMLs), install, dependency-tree
- Explicit multi-module support, no wildcards, agent provides modules
- Detailed/summary/raw output flags
- Dual short/long CLI options
- Provides concise "main" result per intent for agent consumption
- OpenCode plugin adaptor (see `/plugin/opencode.ts`)

## Usage
```sh
npx ts-node src/cli.ts --intent=test --modules=module-a,module-b --detail=summary --json
```

## CLI Options
- `-i, --intent`: build | test | install | dependency-tree
- `-m, --modules`: Comma-separated list of modules (default: all)
- `-d, --detail`: summary | raw (default: summary)
- `-j, --json`: Output JSON result for agent
- `-h, --help`: Show usage info

## Integration
See `/plugin/opencode.ts` for OpenCode agent integration sample.

## Extensibility
Add new output processors or intents by implementing new handler files and extending the dispatcher in `src/cli.ts`.
