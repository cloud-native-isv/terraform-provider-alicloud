## User Input

```text
$ARGUMENTS
```

Process `$ARGUMENTS` per the [User Input Protocol](skills/sdd-workflow/references/user-input-protocol.md). Treat as command parameters, not standalone instructions.

## Checklist Purpose: "Unit Tests for English"

Checklists are **UNIT TESTS FOR REQUIREMENTS WRITING** — they validate quality, clarity, and completeness of requirements. NOT implementation verification.

For detailed methodology, examples, anti-examples, and quality dimension patterns, see `skills/sdd-workflow/references/checklist-methodology.md`.

## Execution Steps

1. **Setup**: Run `.specify/scripts/bash/check-prerequisites.sh --json` from repo root and parse JSON for REQUIREMENTS_DIR and AVAILABLE_DOCS. All paths must be absolute.

2. **Clarify intent**: Derive up to THREE contextual clarifying questions from the user's phrasing + signals from requirements.md. Questions MUST:
   - Be generated from user phrasing + extracted domain signals
   - Only ask about information that materially changes checklist content
   - Be skipped if already unambiguous in `$ARGUMENTS`
   - Follow archetypes: scope refinement, risk prioritization, depth calibration, audience framing, boundary exclusion, scenario class gap
   - Present options as compact table (Option | Candidate | Why It Matters)
   - Defaults: Depth=Standard, Audience=Reviewer, Focus=Top 2 relevance clusters

3. **Understand request**: Combine `$ARGUMENTS` + answers to derive checklist theme, consolidate must-have items, map to category scaffolding.

4. **Load feature context** from REQUIREMENTS_DIR:
   - requirements.md: Feature scope + Requirements (What) + acceptance criteria
   - plan.md (if exists): Specification context for gap-finding only
   - tasks.md (if exists): Specification decomposition for missing requirement detection only

5. **Generate checklist** — Create "Unit Tests for Requirements":
   - Create `REQUIREMENTS_DIR/checklists/` if needed
   - Use short descriptive filename: `[domain].md` (e.g., `ux.md`, `api.md`, `security.md`)
   - Number items sequentially from CHK001
   - Each run creates a NEW file (never overwrites existing)
   - Apply methodology from `skills/sdd-workflow/references/checklist-methodology.md`:
     - Group by quality dimensions (Completeness, Clarity, Consistency, Measurability, Coverage, Edge Cases, Non-Functional, Dependencies, Ambiguities)
     - Each item: question format + quality dimension bracket + traceability reference
     - ≥80% items must include traceability (`[Req §X]`, `[Gap]`, `[Ambiguity]`, etc.)
     - Soft cap: 40 items max, prioritize by risk/impact
   - **PROHIBITED**: Items starting with "Verify/Test/Confirm/Check" + implementation behavior
   - **REQUIRED**: "Are [X] defined/specified?" / "Is [vague term] quantified?" patterns

6. **Structure**: Follow `.specify/templates/checklist-template.md` for canonical format. If unavailable: H1 title, purpose/meta, `##` category sections, `- [ ] CHK### <item>` lines.

7. **Report**: Output path, item count, focus areas, depth level, any user-specified items incorporated.

## Feature Integration

Apply [Feature Integration Protocol](skills/sdd-workflow/references/feature-integration.md). This command transitions status: `Implemented → Ready for Review` (if applicable).

## Handoffs

**Before**: Run after requirements exist (ideally plan/tasks too) so checklist is grounded.

**After**: If items fail → iterate `/speckit.plan` or `/speckit.tasks`. Once satisfied → `/speckit.implement`.