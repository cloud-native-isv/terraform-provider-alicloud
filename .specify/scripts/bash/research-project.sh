#!/usr/bin/env bash

set -e

# Parse command line arguments
JSON_MODE=false
ARGS=()

for arg in "$@"; do
    case "$arg" in
        --json) 
            JSON_MODE=true 
            ;;
        --help|-h) 
            echo "Usage: $0 [--json]"
            echo "  --json    Output results in JSON format"
            echo "  --help    Show this help message"
            exit 0 
            ;;
        *) 
            ARGS+=("$arg") 
            ;;
    esac
done

# Get script directory and load common functions
SCRIPT_DIR="$(CDPATH="" cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/common.sh"

# Ensure UTF-8 locale for better Unicode handling
ensure_utf8_locale || true

# Get all paths and variables from common functions
eval $(get_feature_paths)

# Check if we're on a proper feature branch (only for git repos)
check_feature_branch "$CURRENT_BRANCH" "$HAS_GIT" || exit 1

# Ensure the feature directory exists
mkdir -p "$FEATURE_DIR"

# --- NEW: Research Gathering ---

get_md_title() {
    local file="$1"
    # Read first 5 lines (head -n 5) and look for first header
    local header=$(head -n 5 "$file" | grep -m 1 "^#\{1,\} ")
    if [ -n "$header" ]; then
        # Remove leading hashes and whitespace
        echo "$header" | sed 's/^#\{1,\}[[:space:]]*//'
    else
        echo "(No title)"
    fi
}

# Collect docs
# Find markdown files, limited depth, excluding templates and hidden files
DOC_LIST=""

# Helper to append to DOC_LIST
add_doc() {
    local path="$1"
    if [ -f "$path" ]; then
        local title=$(get_md_title "$path")
        DOC_LIST+="$path: $title"$'\n'
    fi
}

# 1. Project Root Docs
for f in README.md CONTRIBUTING.md CHANGELOG.md; do
    add_doc "$f"
done

# 2. docs/ folder
if [ -d "docs" ]; then
    while IFS= read -r f; do
        add_doc "$f"
    done < <(find docs -maxdepth 2 -name "*.md" | sort)
fi

# 3. Memory (Constitution and Feature Index)
if [ -d ".specify/memory" ]; then
    add_doc ".specify/memory/constitution.md"
    add_doc ".specify/memory/feature-index.md"
fi

# 4. Current Feature Spec (if likely planning)
if [ -f "$FEATURE_SPEC" ]; then
    add_doc "$FEATURE_SPEC"
fi

# Output results
if $JSON_MODE; then
    # Use python to format JSON correctly
    export FEATURE_SPEC IMPL_PLAN FEATURE_DIR CURRENT_BRANCH HAS_GIT DOC_LIST
    python3 -c "
import json, os
data = {
    'FEATURE_SPEC': os.environ.get('FEATURE_SPEC'),
    'IMPL_PLAN': os.environ.get('IMPL_PLAN'),
    'SPECS_DIR': os.environ.get('FEATURE_DIR'),
    'BRANCH': os.environ.get('CURRENT_BRANCH'),
    'HAS_GIT': os.environ.get('HAS_GIT'),
    'AVAILABLE_DOCS': os.environ.get('DOC_LIST')
}
print(json.dumps(data, ensure_ascii=False))
"
else
    echo "FEATURE_SPEC: $FEATURE_SPEC"
    echo "IMPL_PLAN: $IMPL_PLAN" 
    echo "SPECS_DIR: $FEATURE_DIR"
    echo "BRANCH: $CURRENT_BRANCH"
    echo "HAS_GIT: $HAS_GIT"
    echo "AVAILABLE_DOCS:"
    echo "$DOC_LIST"
fi
