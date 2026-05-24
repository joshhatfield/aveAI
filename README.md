
# AVE

## What is AVE?

ave is a lightweight CLI tool for storing and retrieving structured context data that AI agents can query. It fills the gap between "random files in a directory" and "running a full vector database."

## Purpose

AI agents (and humans) need fast, searchable access to project context — code conventions, architectural decisions, notes, patterns, configuration knowledge. Files are slow to search. Full databases are overkill. aveAI is purpose-built for this niche.

## Core Principles

1. **Zero infrastructure** — No servers, no daemons (unless opted in), no Docker. `ave` is a single binary.
2. **Offline-first** — Works without internet. No dependency on embedding APIs for basic search.
3. **Friendly to AI** — CLI output is clean, parseable, and designed to be consumed by both humans and agents.
4. **Progressive enhancement** — Start with text search. Add vectors when you have an embedder. The system degrades gracefully.
5. **Portable** — The `.avdb` file is a single file. Copy it, back it up, ship it with your project.


## Usage

The instal script will allow for global or local opencode install

```shell
## Project specific



curl -fsSL https://raw.githubusercontent.com/joshhatfield/aveAI/master/scripts/install.sh | bash

## Global / all opencode projects 

curl -fsSL https://raw.githubusercontent.com/joshhatfield/aveAI/master/scripts/install.sh | bash -s -- 2

```


## What it is NOT

- Not a general-purpose database
- Not a document store
- Not a replacement for PostgreSQL/SQLite
- Not an MCP server (though one could wrap it)




