## Shell Functions
### check_dir
```bash
check_dir () 
{ 
    [[ -d "$1" && -n $(ls -A "$1" 2>/dev/null) ]] && echo "  ✓ $2" || echo "  ✗ $2"
}
```

### check_feature_branch
```bash
check_feature_branch () 
{ 
    local branch="$1";
    local has_git_repo="$2";
    if [[ "$has_git_repo" != "true" ]]; then
        echo "[specify] Warning: Git repository not detected; skipped branch validation" 1>&2;
        return 0;
    fi;
    if [[ ! "$branch" =~ ^[0-9]+- ]]; then
        echo "ERROR: Not on a feature branch. Current branch: $branch" 1>&2;
```

### check_file
```bash
check_file () 
{ 
    [[ -f "$1" ]] && echo "  ✓ $2" || echo "  ✗ $2"
}
```

### escape_sed
```bash
escape_sed () 
{ 
    echo "$1" | sed 's/[\/&]/\\&/g'
}
```

### find_feature_dir_by_prefix
```bash
find_feature_dir_by_prefix () 
{ 
    local repo_root="$1";
    local branch_name="$2";
    local specs_dir="$repo_root/.specify/specs";
    if [[ ! "$branch_name" =~ ^([0-9]+)- ]]; then
        echo "$specs_dir/$branch_name";
        return;
    fi;
    local prefix="${BASH_REMATCH[1]}";
```

### get_current_branch
```bash
get_current_branch () 
{ 
    if [[ -n "${SPECIFY_FEATURE:-}" ]]; then
        echo "$SPECIFY_FEATURE";
        return;
    fi;
    if git rev-parse --abbrev-ref HEAD > /dev/null 2>&1; then
        git rev-parse --abbrev-ref HEAD;
        return;
    fi;
```

### get_feature_dir
```bash
get_feature_dir () 
{ 
    echo "$1/.specify/specs/$2"
}
```

### get_feature_paths
```bash
get_feature_paths () 
{ 
    local repo_root=$(get_repo_root);
    local current_branch=$(get_current_branch);
    local has_git_repo="false";
    if has_git; then
        has_git_repo="true";
    fi;
    local feature_dir=$(find_feature_dir_by_prefix "$repo_root" "$current_branch");
    cat <<EOF
```

### get_repo_root
```bash
get_repo_root () 
{ 
    if git rev-parse --show-toplevel > /dev/null 2>&1; then
        git rev-parse --show-toplevel;
    else
        local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)";
        ( cd "$script_dir/../../.." && pwd );
    fi
}
```

### has_git
```bash
has_git () 
{ 
    git rev-parse --show-toplevel > /dev/null 2>&1
}
```

### is_valid_utf8
```bash
is_valid_utf8 () 
{ 
    local input="$1";
    if [ -z "$input" ]; then
        return 0;
    fi;
    if printf '%s' "$input" | iconv -f UTF-8 -t UTF-8 > /dev/null 2>&1; then
        return 0;
    else
        echo "Error: Input contains invalid UTF-8 sequences" 1>&2;
```

### safe_quote
```bash
safe_quote () 
{ 
    local input="$1";
    if [ -z "$input" ]; then
        echo "Error: No input provided to safe_quote" 1>&2;
        return 1;
    fi;
    printf '%q' "$input"
}
```

### slugify_unicode
```bash
slugify_unicode () 
{ 
    local input="$*";
    local result="";
    local prev_was_separator=0;
    if [ -z "$input" ]; then
        echo "feature";
        return 0;
    fi;
    local i=0;
```

### validate_input
```bash
validate_input () 
{ 
    local input="$1";
    local max_length="${2:-10000}";
    if [ -z "$input" ]; then
        echo "Error: No input provided to validate_input" 1>&2;
        return 1;
    fi;
    local input_length="${#input}";
    if [ "$input_length" -gt "$max_length" ]; then
```


