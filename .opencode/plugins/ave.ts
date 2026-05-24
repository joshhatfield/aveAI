import type { Plugin } from "@opencode-ai/plugin"
import { tool } from "@opencode-ai/plugin"
import { execSync } from "child_process"

// AveArgs defines all arguments for the ave tool
interface AveArgs {
  command: "search" | "add" | "init" | "list" | "get" | "info" | "context"
  query?: string
  sortKey?: string
  value?: string
  tag?: string
  id?: number
  format?: string
  path?: string
  limit?: number

  // context pull options
  depth?: number      // max hierarchy depth (0 = unlimited)
  counts?: boolean    // show item counts
  summary?: boolean  // summary mode (categories with counts only)
}

const AvePlugin: Plugin = async ({ directory }) => {
  return {
    tool: {
      ave: tool({
        description: `AVE — local context store for AI agents. Store and retrieve project conventions, patterns, notes, and decisions.
Use this tool to:
- search: Find context entries by text query
- add: Store new conventions, patterns, decisions
- init: Initialize a new .avdb database
- list: List entries, optionally filtered by sort-key
- get: Retrieve a single entry by ID
- info: Show database statistics
- context: Pull pseudocontext for LLM seeding (use depth/counts/summary to control output size)

Context pull optimization options (for large databases):
- depth: Limit hierarchy depth to N levels (1 = top-level only, 2 = one nested, etc.)
- counts: Show item counts at each level
- summary: Show only categories with total counts (most compact)`,

        args: {
          command: tool.schema.enum([
            "search", "add", "init", "list", "get", "info", "context"
          ]).describe("The AVE subcommand to execute"),

          // search
          query: tool.schema.string().optional().describe("Search query"),
          sortKey: tool.schema.string().optional().describe("Filter by sort-key prefix (e.g., 'code/conventions')"),
          tag: tool.schema.string().optional().describe("Filter by tag"),
          limit: tool.schema.number().optional().describe("Maximum results (default: 10)"),

          // add
          value: tool.schema.string().optional().describe("The context content to store"),

          // init
          path: tool.schema.string().optional().describe("Database path (default: .ave)"),

          // get
          id: tool.schema.number().optional().describe("Entry ID"),

          // context
          format: tool.schema.string().optional().describe("Output format: 'markdown' or 'json'"),
          depth: tool.schema.number().optional().describe("Max hierarchy depth (0=unlimited, 1=top-level only, etc.)"),
          counts: tool.schema.boolean().optional().describe("Show item counts at each level"),
          summary: tool.schema.boolean().optional().describe("Summary mode - categories with counts only"),
        },

        async execute(args: AveArgs, context: { directory: string }) {
          const cmd = buildAveCommand(args)
          return execAve(cmd, context.directory)
        },
      }),
    },
  }
}

function buildAveCommand(args: AveArgs): string[] {
  const cmd = ["ave", args.command]

  switch (args.command) {
    case "search":
      if (args.query) cmd.push(args.query)
      if (args.sortKey) cmd.push("-s", args.sortKey)
      if (args.tag) cmd.push("-t", args.tag)
      if (args.limit) cmd.push("-l", String(args.limit))
      break

    case "add":
      if (args.sortKey) cmd.push(args.sortKey)
      if (args.value) cmd.push(args.value)
      if (args.tag) cmd.push("-t", args.tag)
      break

    case "init":
      if (args.path) cmd.push(args.path)
      break

    case "list":
      if (args.sortKey) cmd.push(args.sortKey)
      break

    case "get":
      if (args.id !== undefined) cmd.push(String(args.id))
      break

    case "context":
      cmd.push("pull")
      if (args.format) cmd.push("--format", args.format)
      if (args.depth) cmd.push("--depth", String(args.depth))
      if (args.counts) cmd.push("--counts")
      if (args.summary) cmd.push("--summary")
      break

    case "info":
      // no extra args
      break
  }

  cmd.push("--output", "json")
  return cmd
}

function execAve(args: string[], cwd: string): string {
  try {
    const result = execSync(args.join(" "), {
      cwd,
      encoding: "utf-8",
      timeout: 30000,
    })
    return result.trim()
  } catch (e: any) {
    // Return error as JSON so AI can parse it
    return JSON.stringify({
      error: e.message,
      stderr: e.stderr?.toString(),
      code: e.status,
    })
  }
}

export default AvePlugin