---
name: improve-agent
description: General-purpose skill to improve any Spec Kit agent artifact — role template, supervisor, EEI sub-role, orchestration prompt, supervision snippet, or custom agent — from execution feedback, user corrections, and behavioral observations. Use this when the user mentions ["improve agent", "refine agent", "fix agent", "agent feedback", "agent not working", "优化agent", "改进agent", "agent执行反馈"]
skill_id: "<SKILL:.specify/skills/improve-agent/SKILL.md>"
---

# improve-agent

## Goal

Improve **any existing agent artifact** based on evidence from real usage — user feedback, failure cases, behavioral drift, or observed inefficiencies. Targets are not limited to role templates: this skill refines role templates, EEI sub-role templates, orchestration prompts, the shared supervision snippet, and generated custom agents. The result is a targeted update that fixes the identified issues while preserving the artifact's established structure.

## Input Contract

The input is a description of the agent to improve and what went wrong or could be better. Parse:

- **Target identifier**: Resolve to exactly one artifact of a supported kind (see § Target Classification):
  - `templates/agent-role-*-template.md` (role)
  - `templates/agent-subrole-*-template.md` (EEI sub-role)
  - `templates/agent-triad-orchestration-template.md` (orchestration prompt)
  - `templates/agent-supervision-delegation.md` (shared supervision snippet — edits here propagate to ALL supervisors)
  - `.specify/agents/*.agent.md` (a generated custom agent)
- **Improvement direction**: What specifically needs to change — extracted from user feedback, observed failures, or behavioral drift.
- **Evidence**: Concrete examples of the problem (conversation excerpts, incorrect outputs, missing behaviors).

## Target Classification

Before the workflow, classify the target and route to the matching refinement rules:

| Target kind | Match | Route to |
|-------------|-------|----------|
| role | `agent-role-*-template.md` | Workflow steps 1–6 (root-cause on the six mandatory sections) |
| sub-role | `agent-subrole-{executor,evaluator,improver}-template.md` | § Triad Refinement (per-sub-role fixes) |
| orchestration | `agent-triad-orchestration-template.md` | § Triad Refinement (loop/threshold/handoff fixes) |
| supervision snippet | `agent-supervision-delegation.md` | Workflow steps 3–5; WARN that changes affect every supervisor (single source) |
| custom | `.specify/agents/*.agent.md` | Workflow steps 1–6 against the generated file's own structure |

If the identifier matches multiple kinds or none, ask one clarifying question.

## Workflow

### 1. Identify the target template

- Parse the user's input for a role name, slug, or template path
- Resolve to `templates/agent-role-<slug>-template.md`
- If multiple templates match or none match, ask one clarifying question
- Read the current template content before making changes

### 2. Gather evidence

Collect concrete evidence of what needs improvement:

- **User feedback**: Direct statements about what the agent did wrong
- **Behavioral observations**: How the generated agent actually behaved vs. expected behavior
- **Output quality**: Whether the agent's output format matched the template's specification
- **Workflow adherence**: Whether the agent followed its defined workflow steps
- **Handoff issues**: Whether upstream/downstream references worked correctly

### 3. Analyze root causes

For each issue, determine whether the root cause is in:

- **Identity section**: Role definition too vague or too narrow
- **Responsibilities**: Missing duties or conflicting priorities
- **Workflow**: Steps unclear, wrong order, or missing critical steps
- **Upstream/Downstream**: Incorrect references or missing handoff artifacts
- **Output Format**: Expected output not matching what downstream roles need
- **Placeholders**: Wrong context variables for this role's needs

### 4. Apply targeted fixes

- Make minimal, focused changes that address the identified root causes
- Preserve the established template structure (six mandatory sections)
- Do not change sections that are working correctly
- Verify that fixes maintain handoff chain consistency with other roles

### 5. Validate the updated template

- Verify YAML frontmatter still has required fields
- Verify `tools` field remains omitted
- Verify all six mandatory sections are still present
- Verify only approved `{{PLACEHOLDER}}` variables are used
- Verify upstream/downstream references are still consistent

### 6. Report

- List the specific changes made and why
- Reference the evidence that motivated each change
- Suggest re-running `/speckit.agents` to regenerate the agent from the updated template
- Recommend testing the improved agent with the scenario that originally failed

## Constraints

- This skill operates on templates in `templates/`, NOT on generated agents in `.specify/agents/`
- Changes MUST be evidence-based — do not optimize from generic best practices without concrete evidence
- The established template structure (six mandatory sections) MUST be preserved
- Handoff chain consistency with other role templates MUST be maintained
- Prefer minimal changes that fix the observed problem over broad rewrites

## Triad Refinement

Use this workflow when an EEI triad (Executor-Evaluator-Improver) has been running and the user wants to improve sub-role templates based on iteration results.

### What Can Be Refined

- **Executor template**: task interpretation, output format, tool usage patterns
- **Evaluator template**: scoring rubric, acceptance thresholds, dimension weights
- **Improver template**: change strategy, diff granularity, convergence behavior
- **Orchestration prompt**: iteration cap, early-stop criteria, handoff sequencing

### Input

Provide iteration history that includes score progression across rounds and change logs showing what the improver modified at each step. A stalled or oscillating score curve is the clearest signal that one sub-role needs attention.

### Process

1. **Plot the score trajectory** -- identify plateaus, regressions, or oscillations.
2. **Attribute underperformance** -- map each anomaly to the responsible sub-role:
   - Scores plateau early --> Improver generates shallow patches; strengthen its change strategy.
   - Scores oscillate --> Evaluator criteria are ambiguous; tighten rubric definitions.
   - First-round score is very low --> Executor misinterprets the task; clarify its identity/responsibilities.
3. **Draft targeted template edits** using the same minimal-change principle from Step 4 above.
4. **Re-run one iteration** with the updated templates to confirm the fix before committing.

### Examples

| Symptom | Likely Cause | Template Fix |
|---------|-------------|--------------|
| Evaluator misses correctness bugs | Scoring rubric lacks a correctness dimension | Add explicit correctness criteria with weight |
| Improver rewrites instead of patching | Change strategy section too broad | Constrain to diff-level edits, add "preserve working code" rule |
| Executor ignores tool constraints | Identity section missing tool policy | Add tool-usage guardrails to responsibilities |
| Scores converge then regress | Improver over-optimizes one dimension | Add multi-dimension balance check to improver workflow |

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
.specify/memory/feedback/improve-agent-<agent-slug>-<YYYY-MM-DDTHH-MM-SS>.md
```

The feedback document MUST contain:

```markdown
# Agent Execution Feedback

**Source**: improve-agent
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
