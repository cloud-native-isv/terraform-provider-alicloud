---
name: create-agent
description: General-purpose authoring skill for Spec Kit agents — creates role, supervisor (role + embedded EEI triad), triad sub-role, or custom agents. Use this when the user mentions ["create an agent", "new agent", "add agent role", "agent template", "supervisor agent", "EEI triad", "创建agent", "新建agent", "添加角色"]
skill_id: "<SKILL:.specify/skills/create-agent/SKILL.md>"
---

# create-agent

## Goal

Author **any agent artifact** for the Spec Kit agent system — a role template, a role-scoped **supervisor** (role + embedded EEI triad), an EEI **triad** (three sub-role agents + orchestration prompt), or a **custom** `.agent.md`. This skill is the single authoring engine invoked by `/speckit.agents`; the command gathers project context and delegates here rather than rendering templates inline.

## Capability Matrix

Select the capability from the request `kind` (or infer from user intent):

| kind | Produces | Source templates | Primary section |
|------|----------|------------------|-----------------|
| `role` | One role-based agent (six mandatory sections) | `templates/agent-role-*-template.md` | Workflow steps 1–5 below |
| `supervisor` | A role agent that runs its own EEI loop | role template + `agent-supervision-delegation.md` inlined | § Supervisor Capability |
| `triad` | 3 sub-role agents + orchestration prompt | `agent-subrole-*` + `agent-triad-orchestration-template.md` | § Triad Mode (EEI Pattern) |
| `custom` | A single narrow custom `.agent.md` | free-form per intent | `/speckit.agents` Mode B flow |

All capabilities share the same validate + report tail (Workflow steps 4–5) and the Agent-Specific Configuration handling below.

## AgentAuthoringRequest Intake

When invoked by `/speckit.agents`, accept the `AgentAuthoringRequest` defined in `.specify/specs/022-eei-agent-triad/contracts/agent-authoring-contract.md` and consume every field: `kind`, `role_slug`, `task`, `scoring_dimensions[]`, `threshold`, `max_iterations`, `environment_paths[]`, `workspace_paths[]`, `project_context`. Missing optional fields fall back to role/triad defaults (threshold from role, `max_iterations`=20). Return an `AuthoringResult` (`artifact_paths`, `kind`, `status`, `registry_entry`).

## Workflow

### 1. Determine the role definition

**Case A — User provided explicit role description**

Parse the role name, responsibilities, and workflow constraints from the user input:

- **Role name**: A concise display name (e.g., "Security Auditor", "DevOps Engineer")
- **Role slug**: Derive kebab-case slug from role name (e.g., `security-auditor`, `devops-engineer`)
- **Responsibilities**: Core duties of this role in a software development workflow
- **Upstream/Downstream**: Who provides inputs and who consumes outputs

**Case B — User provided no input (empty arguments)**

Analyze the conversation history and project context to infer a useful role:

1. Review conversation for recurring task patterns or specialized workflows
2. Identify the role's position in the development workflow
3. Ask one targeted clarification question if the role is ambiguous

### 2. Validate against existing templates

- Check `templates/agent-role-*-template.md` for existing roles
- If a similar role exists, suggest updating it via `improve-agent` instead
- Ensure the new role does not overlap significantly with the six preset roles

### 3. Create the template file

Write `templates/agent-role-<slug>-template.md` following the established structure:

```markdown
---
name: {{AGENT_NAME}}
description: {{AGENT_DESCRIPTION}}
user-invocable: true
disable-model-invocation: false
---
You are a **<Role Name>** for the {{PROJECT_NAME}} project.

## Identity & Responsibilities
[First-person professional identity and core duties]

## Project Context
[Project-specific placeholders from approved list]

## Workflow
[Step-by-step workflow for this role]

## Upstream (Inputs)
[Who provides inputs and what format]

## Downstream (Outputs)
[Who consumes outputs and what format]

## Output Format
[Expected output structure]
```

### 4. Validate the template

- Verify YAML frontmatter has required fields (name, description, user-invocable)
- Verify `tools` field is omitted (inherits platform defaults)
- Verify all six mandatory sections are present
- Verify only approved `{{PLACEHOLDER}}` variables are used
- Verify upstream/downstream references are consistent with existing role chain

### 5. Report

- Report the created template file path
- Suggest running `/speckit.agents` to generate the new agent from the template
- Propose how this role fits into the existing workflow chain

## Constraints

- Templates MUST follow the established role-based structure (six mandatory sections)
- Templates MUST use only approved `{{PLACEHOLDER}}` variables
- The `tools` field MUST be omitted from YAML frontmatter
- Role instructions MUST be written in first-person professional identity
- This skill operates on templates in `templates/`, NOT on generated agents in `.specify/agents/`

## Triad Mode (EEI Pattern)

Use this mode when the user wants to create an agent that iteratively improves its output through an **Executor-Evaluator-Improver** loop (the "EEI triad").

### When to Use

Trigger on phrases: "create triad", "EEI agent", "iterative quality", "executor evaluator improver", or any request for a self-refining agent that scores and re-drafts its own output.

### How It Works

Instead of producing a single role template, Triad Mode creates **three sub-role agents** plus an **orchestration prompt** that wires them into a loop:

1. **Executor** -- performs the core task and produces a draft artifact.
2. **Evaluator** -- scores the draft against weighted dimensions, emits a structured rubric, and decides pass/fail against a threshold.
3. **Improver** -- reads the rubric, rewrites the artifact to address low-scoring dimensions, and feeds the revision back to the Evaluator.

The orchestration prompt drives the loop: Executor -> Evaluator -> (if below threshold) Improver -> Evaluator -> ... until the threshold is met or a max-iteration cap is reached.

### Required Inputs

| Input | Description |
|-------|-------------|
| **Task description** | What the Executor should produce (e.g., "generate an API spec") |
| **Scoring dimensions + weights** | Named quality axes and their relative weights (e.g., correctness 0.4, completeness 0.3, clarity 0.3) |
| **Threshold** | Minimum weighted score (0-1) for the Evaluator to accept the artifact |
| **Environment paths** | Project root, templates dir, output dir |
| **Workspace paths** | Where intermediate artifacts and rubrics are written |

### Template Files

All four templates live in `templates/`:

- `agent-subrole-executor-template.md`
- `agent-subrole-evaluator-template.md`
- `agent-subrole-improver-template.md`
- `agent-triad-orchestration-template.md`

The skill populates placeholders and writes the generated files to `.specify/agents/`.

### Triad Workflow (within this skill)

1. Collect required inputs (prompt or infer from conversation).
2. Validate that all four subrole templates exist in `templates/`.
3. Generate three sub-role agent files and one orchestration prompt.
4. Report the created file paths and suggest a test invocation.

## Supervisor Capability

Use this capability (`kind: supervisor`) to author a **role-scoped supervisor** — a role agent that, for quality-gated deliverables, orchestrates its own EEI loop. Per the amendment decisions: supervision is **active by default** (OQ-1) and the delegation section is **composed at generation, not copied into role templates** (OQ-2).

### How it works

1. Load the role template `templates/agent-role-<role_slug>-template.md` (it carries `supervisor: true` and `role-scope: <role_slug>` in frontmatter).
2. **Inline the single-source snippet** `templates/agent-supervision-delegation.md` into the generated agent body, resolving its placeholders:
   - `{{ROLE_SCOPE}}` → `role_slug`
   - `{{ROLE_NAME}}` → role display name
   - `{{ROLE_DIMENSIONS}}` → the request's `scoring_dimensions` (or role-appropriate defaults if omitted)
3. Preserve `supervisor: true` in the generated frontmatter (honor `supervisor: false` only if the request explicitly opts out).
4. Write the generated agent to `.specify/agents/<role_slug>.agent.md` and run the shared validate + report tail.

### Rule

Never copy the delegation section text into `templates/agent-role-*-template.md`; edit it only in `templates/agent-supervision-delegation.md` so all supervisors stay in sync (single source of truth).

## Agent-Specific Configuration

### Step 1: Identify Executing Agent

Before executing this skill's workflow, identify which AI agent you are:

| Agent | Detection Signals |
|-------|-------------------|
| **Claude Code** | System prompt contains "Claude Code"; tools include `Agent`, `Edit`, `Bash`, `Read`; `.claude/` directory exists |
| **GitHub Copilot** | Running in VS Code Copilot Chat context; `.github/copilot-instructions.md` loaded; tools include `workspace edit`, `@terminal` |
| **Qoder CLI** | `.qoder/` directory exists; `AGENTS.md` instructions loaded |
| **opencode** | `.opencode/` directory exists |
| **Qwen Code** | `QWEN.md` instructions loaded; `.qwen/` directory exists |
| **Codex CLI** | `.codex/` directory exists |
| **Hermes Agent** | `.hermes/` directory exists |
| **iFlow** | `.iflow/` directory exists |

If you cannot identify your agent, skip Step 2 and proceed with the standard workflow.

### Step 2: Load Agent-Specific Guidance

If you identified your agent in Step 1, check if a guide exists at:

```
${SKILL_HOME}/references/<agent-slug>-guide.md
```

Where `<agent-slug>` is: `claude-code`, `copilot`, `qoder`, `opencode`, `qwen`, `codex`, `hermes`, or `iflow`.

If the guide exists, read it and apply the agent-specific tool mappings, best practices, and pitfall avoidances during execution. If no guide exists for your agent, proceed with the standard workflow.

### Step 3: Capture Execution Feedback

If you encounter an agent-specific obstacle during execution (e.g., a tool call is unavailable, output format doesn't match expectations, a workaround was needed), generate a feedback document at:

```
.specify/memory/feedback/create-agent-<agent-slug>-<YYYY-MM-DDTHH-MM-SS>.md
```

The feedback document MUST contain:

```markdown
# Agent Execution Feedback

**Source**: create-agent
**Agent**: <agent-slug>
**Timestamp**: <ISO-8601>
**Outcome**: <success-with-workaround | partial-failure | full-failure>

## Obstacle
[Description of the agent-specific issue encountered]

## Workaround Applied
[What was done to work around the issue, if anything]

## Suggested Improvement
[Specific change to the skill or reference document that would prevent this issue]
```

Only generate feedback when a genuine agent-specific obstacle was encountered.
