## User Input

```text
$ARGUMENTS
```

You **MUST** consider the user input before proceeding (if not empty).

## Outline

You are updating the project feature index at `.specify/memory/features.md`. This file is a TEMPLATE containing placeholder tokens in square brackets (e.g. `[PROJECT_NAME]`, `[FEATURE_COUNT]`). Your job is to (a) collect/derive concrete values, (b) fill the template precisely, and (c) ensure proper feature entry formatting.

Follow this execution flow:

1. Load the existing feature index template at `.specify/memory/features.md`.
   - Identify every placeholder token of the form `[ALL_CAPS_IDENTIFIER]`.
   - **IMPORTANT**: The user might provide new features, updates to existing features, or general project context. Parse the input accordingly.

2. Collect/derive values for placeholders:
   - If user input (conversation) supplies feature descriptions or project context, use it to generate/update features.
   - Otherwise infer from existing repo context (README, docs, existing spec directories if present).
   - For dates: `LAST_UPDATED_DATE` and individual feature `FEATURE_LAST_UPDATED` should be today's date (YYYY-MM-DD format).
   - `FEATURE_COUNT` should reflect the actual number of features after processing user input.
   - `PROJECT_NAME` should be derived from repository name, README title, or user input.

3. Process feature entries:
   - Parse any existing feature entries from `.specify/specs/` directories if they exist
   - Generate new feature entries from user input with sequential IDs (001, 002, etc.)
   - Each feature entry must include: ID, Name (2-4 words), Description, Status, Spec Path, Last Updated
   - Status options: Draft, Planned, Implemented, Ready for Review, Completed
   - Spec Path should point to actual spec file path or "(Not yet created)" for new features

4. Draft the updated feature index content:
   - Replace every placeholder with concrete text (no bracketed tokens left)
   - Format `[FEATURE_ENTRIES]` as a proper Markdown table with all feature entries
   - Preserve heading hierarchy and remove template instruction comments once replaced
   - Ensure table columns match the defined format exactly

5. Validation before final output:
   - No remaining unexplained bracket tokens
   - All dates in ISO format YYYY-MM-DD
   - Feature IDs are sequential three-digit numbers
   - Status values are from the allowed set
   - Table structure is valid Markdown

6. Write the completed feature index back to `.specify/memory/features.md` (overwrite).

7. Output a final summary to the user with:
   - Number of features processed/created
   - Any new feature IDs generated
   - Suggested commit message (e.g., `docs: update feature index with new entries`)

Formatting & Style Requirements:

- Use Markdown headings exactly as in the template
- Table must have proper alignment with header row
- Keep a single blank line between sections
- Avoid trailing whitespace
- Feature names should be concise (2-4 words maximum)

If the user supplies partial updates (e.g., only one new feature), merge with existing features and update accordingly.

If critical info missing (e.g., project name unknown), use repository name or insert reasonable default.

Do not create a new template; always operate on the existing `.specify/memory/features.md` file.