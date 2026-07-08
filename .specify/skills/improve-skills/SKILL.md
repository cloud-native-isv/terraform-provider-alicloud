---
name: improve-skills
description: This skill continuously improves one local Skill from a user-provided Skill description, execution history, user feedback, failure cases, and observed inefficiencies. Use this when the user mentions ["improve skills after use", "skill execution feedback", "refine SKILL.md", "skill retrospective", "skill iteration", "技能执行反馈", "基于执行问题优化skill", "持续改进Skill"]
skill_id: "<SKILL:.specify/skills/improve-skills/SKILL.md>"
---

# improve-skills

## Goal

Continuously improve one existing local SpecKit Skill from a user-provided Skill description and evidence from real executions. The expected result is a focused Skill update that fixes observed problems, captures reusable lessons, and makes the next execution more reliable.

## Input Contract

The input is a description of the Skill to improve. It must be interpreted as follows:

- **Target identifier**: identify exactly one local Skill by `skill_id`, frontmatter `name`, canonical path, or Skill directory name. If multiple Skills match or no local Skill can be found, ask one targeted clarification before editing.
- **Optimization direction**: extract the requested direction when present, such as fixing execution failures, improving efficiency, clarifying inputs/outputs, correcting tool usage, or strengthening validation. If no direction is present, infer it only from concrete execution history and user feedback.
- **User emphasis**: treat details in the user's description as high-priority evidence. Analyze them explicitly even when broader execution history suggests additional improvements.

## Workflow

1. **Identify the target Skill, optimization direction, and execution window**
   - Parse the user's Skill description for a `skill_id`, frontmatter `name`, canonical path, or Skill directory name; improve only that local Skill.
   - If the user says “this Skill”, infer the target from the active file or recent conversation, then verify it resolves to exactly one local Skill.
   - Extract the requested optimization direction when present and carry it through evidence collection, analysis, edits, validation, and reporting.
   - Treat `.specify/skills/<name>/SKILL.md` as the canonical source of truth; use `.github/skills/<name>` only as a compatibility entrypoint.
   - Re-read the canonical `SKILL.md` before editing, especially when the system reports recent user or formatter changes, or when a refresh/script may have modified metadata.
   - Define the execution window to review: current conversation, last Skill run, failed command output, user correction, test failure, or recent edits.
   - When improving `improve-skills` itself, use the most recent improvement loop as the execution window and avoid reapplying the same lesson unless new evidence shows the previous fix was insufficient.

2. **Measure execution effectiveness from history**
   - Gather concrete evidence before editing: user feedback, steps that were confusing, tool failures, wrong assumptions, repeated manual fixes, validation gaps, and changed files from the execution.
   - Include terminal/test outputs and error messages when they explain what went wrong.
   - Review changed files as evidence, but classify generated validation artifacts such as `tools/*.json` separately from hand-edited Skill instructions.
   - Measure the target Skill against the requested or inferred optimization goal: whether it could be invoked, whether its expected input format was accepted, whether the workflow produced the expected output, how many avoidable manual/tool steps occurred, and whether validation caught the issue.
   - Identify the execution-flow steps that did not meet expectations, including broken command-line parameters, mismatched expected formats, missing prerequisites, ambiguous target resolution, inefficient tool choices, repeated searches, or unnecessary user handoffs.
   - Separate facts from interpretation. Do not optimize from generic best-practice principles when no execution evidence supports the change.
   - **Treat executor-reported runtime failures as first-class evidence, outranking any prior "it works" claim.** A reference/example snippet that throws or silently no-ops at runtime (missing module, a loop that toggles its own state, an extractor that skips iframes) is a confirmed defect even if the file exists and a previous run asserted the Skill was fine. Anchor the fix to the observed stack trace / wrong output, not to the earlier assertion; a stale "already works" is itself a finding to correct.
   - **Silent under-extraction counts as a defect, not just a throw.** When an executor reports a helper "ran fine" but returned *empty or thin* results (a select with no options, a dashboard with "0 panels", `variables: []` on a page that visibly has them), that is evidence the helper read the wrong/too-early DOM — not evidence the page is empty. Capture the executor's concrete observation (which selector matched 0 nodes, which content was missing) as the reusable fact; it is often more actionable than any error message because nothing crashed. Record it even when the run's headline metrics (coverage, screenshots) all look green.
   - If evidence is insufficient, ask one targeted question about what failed, what was inefficient, or what should happen differently next time.

3. **Analyze user-provided emphasis and organize improvement items**
   - Give the user's stated optimization direction a dedicated analysis pass: confirm which parts are already satisfied, which parts are missing, and which edits will directly address the request.
   - Group observations by failure mode: trigger/discovery, scope inference, missing context, wrong tool choice, unsafe step, unclear output, validation gap, or resource/reference issue.
   - For each item, record: observed symptom, likely cause in the current Skill instructions, desired next behavior, and the file section to change.
   - Discard one-off environment noise unless the Skill should explicitly handle it in future runs. If a refresh command exits successfully with a fallback after an optional source warning, record it as a validation note rather than a root cause.
   - **Legacy path idioms**: when the Skill under review still uses any of the following, flag them as migration candidates and apply the Migration Mapping table from `templates/commands/skills.md` (`## Migration Mapping`):
     - Bare relative paths such as `./scripts/init.sh` or `./references/checklist.md` → rewrite as `${SKILL_HOME}/...`.
     - `${SKILL_ROOT}/X` references → rewrite as `${SKILL_HOME}/X`.
     - Agent-specific install paths embedded in prose (e.g., `~/.copilot/skills/<name>/...`, hard-coded `.specify/skills/<name>/...`) → rewrite as `${SKILL_HOME}/...`.

4. **Correct the root causes with minimal changes**
   - For complete execution failures, fix the instruction that caused non-execution first, such as wrong command-line arguments, nonexistent paths, invalid expected file formats, incompatible metadata, or missing prerequisite checks.
   - For successful but inefficient executions, replace the inefficient step with a more direct method, deterministic script, narrower search, better evidence filter, or clearer decision branch.
   - Prefer changing the step that caused the observed problem over adding broad new rules.
   - Convert repeated user corrections into explicit decision branches.
   - Convert repeated manual checks into checklist items or deterministic scripts when appropriate.
   - **Harden reference helpers that read a live third-party/framework DOM against the two recurring silent-under-extraction causes.** (1) *Interaction-gated content* — options, tab bodies, popovers, and menus that mount only on click (e.g. a component library's `<Select>` whose options render in a portal) cannot be read from a static-DOM scrape; the fix is a bounded open-then-read-then-close pass at the driver level, not a richer static selector. (2) *Version-drifting selectors* — third-party embeds (dashboards, editors) rename classes/`data-testid`s between versions, so a single selector silently matches 0 nodes; query the old **and** current shapes and treat an empty result on a page known to have the element as a drift signal to re-check, not as absence. Prefer these targeted robustness edits over widening the whole extractor. For concrete before/after code (the brittle version, the runtime symptom, the hardened version, and the general lesson) see [`./references/hardening-examples.md`](./references/hardening-examples.md).
   - **Prefer ground-truth signals over convenient-but-misleading ones, and scope extraction to the intended node.** A count or label the target system already prints (a "(0 panels)" row header, a "3 results" badge, a status pill) looks authoritative but is often a stale/partial/collapsed-state artifact — do not surface it as the fact. Derive the value from the underlying elements instead (count the real DOM nodes, read the actual list), and reconcile: when the emitted label disagrees with the element-level truth, treat the elements as authoritative and stop reporting the label. Likewise, when a broad selector's text is used as a field, confirm it captures only that field — a container/ancestor selector silently absorbs sibling/body content (a title jammed onto a whole table row, a name jammed onto its value). The fix is to scope the selector to the leaf/header node, not to post-filter the polluted string (length caps are a backstop, not the fix). Mislabeling one datum as another (a selected value reported as the variable's name) is the same class of error — verify each captured field is the thing the doc claims it is.
   - Move detailed lessons to `./references/` only when they are useful but not needed every run.
   - **Slim `SKILL.md` toward contract-only content**: when an edit touches `SKILL.md`, also evaluate whether existing sections should be moved out per [`./references/skill-slimming-principles.md`](./references/skill-slimming-principles.md). The body is a contract — frontmatter, resource index, workflow skeleton, strict requirements, and conventions — not a manual. How-to checklists, error tables, command-pair comparisons, environment-detection scripts, install commands, and intra-domain routing tables belong in references; replace them with a one-sentence pointer + anchor link. Always **delete-and-absorb** (copy the substantive content into the target reference in the same edit), never delete-and-drop. Defer environment-level recovery (auto-install, shell switching, OS branching) to the user — surface the error and fix command, then stop. **Never slim a section that a feature spec or contract test mandates be present inline.** Before moving/removing any named section (heading) out of `SKILL.md`, grep `.specify/specs/**` and `tests/contract/**` for that heading; if a contract asserts its inline presence (e.g. `## Agent-Specific Configuration` with `### Step 1/2/3` per `021-agent-specific-config` contract C-002), it is part of the skill's contract — keep it inline and slim elsewhere. Structural moves that a heading-grep test can see are the highest-risk slimming edits.

5. **Update the Skill for the next execution**
   - Edit `SKILL.md` to make the improved behavior executable and checkable.
   - Update frontmatter `description` only when execution feedback shows trigger/discovery mismatch.
   - Update `./references/`, `./scripts/`, or `./assets/` only when the evidence shows they will reduce future mistakes.
   - Avoid adding process logs, changelogs, or full retrospectives to the Skill; distill only reusable lessons.

6. **Validate the improvement loop**
   - Re-read the changed Skill and verify that each edit maps to an observed execution issue.
   - **When an edit touches reference code or example snippets, validate that the code actually RUNS — not merely that the file exists or parses.** Execute it (or the smallest reproducing harness) against a real target, or trace it line-by-line against the documented API to confirm the control flow does what the prose claims. Files-exist / links-resolve checks do not catch a snippet that throws, loops incorrectly, or no-ops; those surface only at runtime. If you cannot run it this loop, say so and mark it as needing runtime validation in the next real execution rather than reporting it as verified.
   - Check frontmatter, resource paths, line count, compatibility entry, and registry row when metadata changed. If `skill_id` is added or corrected, ensure `.specify/instructions.md` has one deduplicated Skills registry row for the canonical Skill.
   - Accept a directory-level `.github/skills -> ../.specify/skills` symlink as a valid compatibility entrypoint; do not require a separate per-Skill symlink when the directory symlink already exposes the Skill.
   - **After any structural edit to a SKILL.md or its references (moving/renaming/removing a section or file), run the affected contract tests** (`tests/contract/` for the feature that governs the skill) — not just grep for headings. A slimming move can silently break a heading-presence contract test; only executing the tests proves it still passes.
   - If a combined validation command returns only partial output or omits later checks, rerun the missing checks individually before concluding validation passed.
   - Do not document `.specify/scripts/` as a Skill-owned resource directory; Skill-owned executable resources belong in `./scripts/`.

7. **Report the feedback-driven changes**
   - Summarize the execution feedback that drove the update.
   - List changed Skill files and the behavior expected to improve next time.
   - Note any unresolved feedback that needs another real execution to validate.

## Quality Checklist

Use [the Skill quality checklist](./references/skill-quality-checklist.md) to structure execution feedback, root-cause analysis, and validation when the improvement involves more than one observed issue.

## Resource ID

- Canonical ID: `<SKILL:.specify/skills/improve-skills/SKILL.md>`
- Canonical Path: `.specify/skills/improve-skills/SKILL.md`

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
.specify/memory/feedback/improve-skills-<agent-slug>-<YYYY-MM-DDTHH-MM-SS>.md
```

The feedback document MUST contain:

```markdown
# Agent Execution Feedback

**Source**: improve-skills
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

## Resources

| Directory | Contents |
|-----------|----------|
| `${SKILL_HOME}/references/` | `skill-slimming-principles.md`, `skill-quality-checklist.md`, `hardening-examples.md`, `claude-code-guide.md`, `copilot-guide.md` |
