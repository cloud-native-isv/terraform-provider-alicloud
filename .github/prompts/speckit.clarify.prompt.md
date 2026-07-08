## User Input

```text
$ARGUMENTS
```

Process `$ARGUMENTS` per the [User Input Protocol](skills/sdd-workflow/references/user-input-protocol.md). Treat as command parameters, not standalone instructions.

## Outline

Goal: Detect and reduce ambiguity or missing decision points in the current phase's primary artifact and record clarifications directly in that file.

### Phase 0: Phase Detection & Context Setup

1. Run `.specify/scripts/bash/check-prerequisites.sh --json --paths-only` from repo root once; parse JSON for `REPO_ROOT`, `BRANCH`, `REQUIREMENT_ID`, `REQUIREMENTS_DIR`, `FEATURE_SPEC`, `IMPL_PLAN`, `TASKS`.

2. **Determine current phase** by file existence in REQUIREMENTS_DIR:

   | Phase | Trigger | Target File | Handoff |
   |-------|---------|-------------|---------|
   | A: Post-Requirements | `plan.md` NOT exist | `requirements.md` | `/speckit.plan` |
   | B: Post-Plan | `plan.md` exists, `tasks.md` NOT exist | `plan.md` | `/speckit.tasks` |
   | C: Post-Tasks | `tasks.md` exists | `tasks.md` | `/speckit.implement` |

   - If no `requirements.md` at all → abort, instruct `/speckit.requirements` first.
   - Log: `**Mode: [A/B/C] — clarifying [target]**`

3. **Load common context**: `.specify/memory/constitution.md`, `README.md`, relevant `docs/`, `.specify/memory/features.md`, `research.md` (if exists).

4. **Load mode-specific context** (see taxonomy reference).

### Taxonomy & Coverage Scan

For the active mode's detailed taxonomy categories and integration rules, load: `skills/sdd-workflow/references/clarify-taxonomy.md`

Each mode has its own taxonomy. For each category, mark status (Clear / Partial / Missing). Add candidate questions for Partial/Missing categories unless clarification would not materially change implementation.

If spec contains `Feature ID: Need clarification` or `Feature Name: Need clarification`, treat Feature Linkage as high-priority.

### Question Generation & Interactive Loop

1. **Generate prioritized queue** (max 5 questions internally). Constraints:
   - Filter via Research: Skip questions answered by `research.md`
   - Maximum 10 total across session
   - Each must be answerable with multiple-choice (2–5 options) OR ≤5-word short answer
   - Only include questions whose answers materially impact downstream artifacts
   - Balance category coverage; favor high-impact unresolved categories

2. **Sequential questioning loop** — present ONE question at a time:
   - Multiple-choice: table format (Option | Description), state **Recommended** option with reasoning
   - Short-answer: state **Suggested** answer with reasoning
   - User replies: "yes"/"recommended" → use suggestion; otherwise validate answer
   - Stop when: all critical ambiguities resolved, user signals "done", or 5 questions reached

3. **Integration** — use the active mode's rules from the taxonomy reference.

4. **Validation** (after each write):
   - One bullet per accepted answer in `## Clarifications` > `### Session YYYY-MM-DD`
   - No lingering vague placeholders remain
   - Markdown structure valid; terminology consistent

5. **Final Write**: Save the target file.

### Report

- Operating mode (A/B/C) and target file path
- Questions asked & answered count
- Sections touched
- Coverage summary table (Resolved / Deferred / Clear / Outstanding)
- Suggested next command: A → `/speckit.plan`, B → `/speckit.tasks`, C → `/speckit.implement`

## Behavior Rules

- If no meaningful ambiguities: "No critical ambiguities detected." → suggest proceeding
- Never exceed 5 total questions
- Respect early termination signals ("stop", "done", "proceed")
- Do not modify upstream artifacts in later modes (B won't edit requirements.md; C won't edit plan.md)

## Handoffs

**Before**: At minimum `/speckit.requirements` first. Phase auto-detected from file existence.

**After**: Mode A → `/speckit.plan`. Mode B → `/speckit.tasks`. Mode C → `/speckit.implement`.