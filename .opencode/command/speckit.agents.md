> Compatibility: Follow VS Code Copilot custom agent format for `.agent.md` files.

## User Input

```text
$ARGUMENTS
```

Process `$ARGUMENTS` per the [User Input Protocol](skills/sdd-workflow/references/user-input-protocol.md). If empty, infer agent intent from conversation/repo context. If intent unclear, ask concise clarification.

## Outline

Goal: Generate role-based agents or create custom agents at `.specify/agents/<name>.agent.md` (canonical location). Tool-specific directories are symlinks — do NOT write to them directly.

### Delegation Model

This command does NOT render templates inline. Both modes delegate to `create-agent` skill (new agents) or `improve-agent` skill (updates). The command gathers context, builds the `AgentAuthoringRequest`, handles backup/preservation, verifies symlinks, and updates registry.

### Mode A: Role-Based Generation (no arguments)

Generate all six role agents from templates in `.specify/templates/agent-role-*-template.md`:

| Role | Template | Output |
|------|----------|--------|
| Requirements Analyst | `agent-role-requirements-analyst-template.md` | `requirements-analyst.agent.md` |
| System Designer | `agent-role-system-designer-template.md` | `system-designer.agent.md` |
| Module Designer | `agent-role-module-designer-template.md` | `module-designer.agent.md` |
| Test Engineer | `agent-role-test-engineer-template.md` | `test-engineer.agent.md` |
| QA Engineer | `agent-role-qa-engineer-template.md` | `qa-engineer.agent.md` |
| Knowledge Manager | `agent-role-knowledge-manager-template.md` | `knowledge-manager.agent.md` |

**Flow**: Gather project context (README, pyproject.toml, constitution, features, specs) → For each role, invoke `create-agent` skill with `kind: supervisor` → Backup existing if customized (FR-008a) → Preserve non-role agents (FR-008) → Write to `.specify/agents/` → Scaffold workspace files if needed (AGENTS.md, MEMORY.md, SOUL.md, USER.md) → Report.

### Mode B: Custom Agent (with arguments)

Route to `create-agent` (new, `kind: custom`) or `improve-agent` (existing). For EEI triad pass `kind: triad`, for role supervisor pass `kind: supervisor`.

**Flow**: Determine intent → Check `.specify/agents/<name>.agent.md` existence → Extract from conversation (role, tools, domain) → Define agent shape (name, description, tools, model, handoffs) → Delegate to skill → Write file → Register `agent_id` in `.specify/instructions.md` → Verify symlinks → Report.

### Authoring Rules

- Focus on **what** and **when to call** the agent
- Concise, explicit instructions over narrative
- Single responsibility per agent
- Least-privilege tool set
- Approved providers: Claude Code, GitHub Copilot, Qwen Code, opencode, Qoder

### Frontmatter Baseline

```yaml
---
description: "<required: trigger words + when to use>"
tools: ["read", "search"]
---
```

Supported fields: `name`, `tools`, `model`, `argument-hint`, `agents`, `user-invocable`, `disable-model-invocation`, `handoffs`.

### Valid File Locations

- Canonical: `.specify/agents/*.agent.md`
- Symlinks (read-only): `.github/agents/`, `.qoder/agents/`, `.qwen/agents/`, `.opencode/agents/`

### Validation

- YAML frontmatter must be valid
- Reject unsupported provider references
- Tool list must match workflow needs
- Unresolved contradictions block save

For agent-specific operational guidance, see `skills/sdd-workflow/references/agent-configuration.md`.

## Handoffs

**Before**: Optional `/speckit.skills` if agent depends on new skill. Optional `/speckit.tools` for tool records.

**After**: Run `/speckit.instructions` to sync discoverability.