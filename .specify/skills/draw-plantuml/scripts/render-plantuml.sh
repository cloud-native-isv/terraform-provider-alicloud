#!/usr/bin/env bash
# render-plantuml.sh — Render PlantUML to high-quality SVG and adaptive PNG
#
# SVG: Always rendered at scale 4 + dpi 300 (vector, no size limit).
# PNG: Adaptively calculates scale/dpi to fit within PNG_MAX (default 4095),
#      ensuring output is never blank due to server buffer overflow.
#
# The PlantUML server has a hard PNG dimension cap of 4096×4096. When the
# internal rendering buffer exceeds this, it silently returns a blank image.
# This script targets PNG_MAX (4095) to stay safely below the cap.
#
# Two rendering backends (auto-selected):
#   • server — POST to a PlantUML server (PLANTUML_SERVER). Preferred when reachable.
#   • local  — a local PlantUML jar (PLANTUML_JAR or a well-known path) via `java`.
#             Works fully offline. Required diagram types that need Graphviz
#             (class/component/deployment/sequence/state/usecase/activity/package)
#             still need `dot` on PATH; WBS/Gantt/MindMap/JSON/YAML do not.
# Backend selection: PLANTUML_BACKEND=server|local|auto (default auto).
#   auto → probe the server; on failure fall back to the local jar if present.
#
# Usage: render-plantuml.sh <input.puml> [output_dir] [output_prefix]

set -euo pipefail

PLANTUML_SERVER="${PLANTUML_SERVER:-http://workspace.code-workspace.cloud:39156/plantuml}"
PLANTUML_BACKEND="${PLANTUML_BACKEND:-auto}"   # server | local | auto
SVG_SCALE=4        # SVG: maximum quality (vector, no size limit)
SVG_DPI=300
PNG_MAX=4095       # PNG: target max dimension (< server hard cap 4096)
PNG_DPI=300        # PNG: preferred DPI for high pixel density
PNG_MIN_DPI=96     # PNG: minimum DPI fallback
PNG_BLANK_THRESHOLD=100000  # PNG file size below this for 4096×4096 = likely blank

log() { printf '[render-plantuml] %s\n' "$*" >&2; }
warn() { printf '[render-plantuml] WARNING: %s\n' "$*" >&2; }

# ── Backend resolution ────────────────────────────────────────────────────────

# Resolve a local PlantUML jar path. Order: $PLANTUML_JAR, well-known locations.
resolve_jar() {
  if [[ -n "${PLANTUML_JAR:-}" ]] && [[ -f "$PLANTUML_JAR" ]]; then
    echo "$PLANTUML_JAR"; return 0
  fi
  local cand
  for cand in \
    "$HOME/.local/share/plantuml/plantuml.jar" \
    "/usr/local/share/plantuml/plantuml.jar" \
    "/opt/plantuml/plantuml.jar" \
    "/tmp/plantuml.jar"; do
    [[ -f "$cand" ]] && { echo "$cand"; return 0; }
  done
  return 1
}

# Probe whether the PlantUML server can render (POST a trivial diagram).
server_reachable() {
  local probe='@startuml
a->b
@enduml'
  curl -sf -m 6 -X POST -H "Content-Type: text/plain" \
    --data-binary "$probe" "${PLANTUML_SERVER}/svg" -o /dev/null 2>/dev/null
}

# Decide which backend to use. Echoes "server" or "local"; exits on neither.
select_backend() {
  local jar; jar="$(resolve_jar || true)"
  case "$PLANTUML_BACKEND" in
    server) echo "server" ;;
    local)
      [[ -n "$jar" ]] || { warn "PLANTUML_BACKEND=local but no jar found"; exit 1; }
      echo "local" ;;
    auto|*)
      if server_reachable; then
        echo "server"
      elif [[ -n "$jar" ]]; then
        warn "Server unreachable; using local jar: ${jar}"
        echo "local"
      else
        warn "PlantUML server unreachable and no local jar found."
        warn "Set PLANTUML_JAR=/path/to/plantuml.jar or start a server."
        exit 1
      fi ;;
  esac
}

# Render a styled .puml to a target format via the chosen backend.
# render_diagram <styled_puml> <out_file> <svg|png>
render_diagram() {
  local styled="$1" out="$2" fmt="$3"
  if [[ "$BACKEND" == "local" ]]; then
    local outdir; outdir="$(cd "$(dirname "$out")" && pwd)"
    local base; base="$(basename "$styled")"; base="${base%.puml}"
    # PlantUML writes <base>.<ext> into -o dir; capture stderr for diagnostics.
    # PLANTUML_LIMIT_SIZE lifts the default 4096px cap so high-scale/high-dpi
    # renders are not silently clamped (local jar has no server-side limit).
    if ! java -Djava.awt.headless=true -DPLANTUML_LIMIT_SIZE="${PLANTUML_LIMIT_SIZE:-16384}" \
          -jar "$JAR" "-t${fmt}" -charset UTF-8 \
          -o "$outdir" "$styled" >/dev/null 2>"${out}.jarlog"; then
      warn "Local jar rendering failed (${fmt}):"; sed 's/^/[jar] /' "${out}.jarlog" >&2 || true
      rm -f "${out}.jarlog"; return 1
    fi
    rm -f "${out}.jarlog"
    local produced="${outdir}/${base}.${fmt}"
    [[ "$produced" != "$out" ]] && mv -f "$produced" "$out"
    [[ -f "$out" ]]
  else
    curl -sf -X POST -H "Content-Type: text/plain" \
      --data-binary "@${styled}" "${PLANTUML_SERVER}/${fmt}" -o "$out"
  fi
}

# ── Style injection ───────────────────────────────────────────────────────────

# Remove existing style directives that we'll inject (avoid duplicates).
# Direction directives (top to bottom, left to right) are preserved.
strip_style() {
  local input="$1"
  sed -E \
    -e '/^[[:space:]]*skinparam[[:space:]]+(monochrome|shadowing|roundCorner|dpi|defaultFontSize|defaultFontName|padding|ArrowThickness|BorderThickness|svgDimensionStyle|svgLinkTarget|actorStyle)[[:space:]]/d' \
    -e '/^[[:space:]]*scale[[:space:]]/d' \
    "$input"
}

# Inject style block with given scale and dpi after the diagram's @start tag.
inject_style() {
  local input="$1" output="$2" scale="$3" dpi="$4"
  local style_tmp="${output}.style.tmp"

  # Detect the actual start tag (@startuml / @startwbs / @startgantt / ...).
  # The style block must be inserted right after it, not hardcoded to @startuml.
  local start_tag
  start_tag=$(grep -m1 -oiE '^[[:space:]]*@start[a-z]+' "$input" 2>/dev/null | tr -d '[:space:]')
  [[ -z "$start_tag" ]] && start_tag="@startuml"

  # Is this a specialty (non-UML) diagram?
  local specialty=""
  if [[ "$start_tag" =~ ^@start(wbs|gantt|mindmap|json|yaml|salt)$ ]]; then
    specialty="true"
  fi

  # Detect if source uses color markup — skip monochrome if so
  local use_mono="true"
  if grep -qE '<color:|<font color|skinparam monochrome false' "$input" 2>/dev/null; then
    use_mono=""
    log "Color markup detected — skipping monochrome"
  fi

  if [[ -n "$specialty" ]]; then
    # Specialty diagrams (WBS/Gantt/MindMap/JSON/YAML/Salt) rely on native
    # coloring and their own <style> blocks. Inject only scale + dpi (so the
    # image is rendered large & crisp instead of raw 1:1) + a CJK-capable font.
    # Do NOT force monochrome or UML skinparams here.
    log "Specialty diagram (${start_tag}) — minimal style: scale=${scale}, dpi=${dpi}"
    cat <<EOF > "$style_tmp"
skinparam dpi ${dpi}
scale ${scale}
skinparam shadowing false
skinparam defaultFontName "Noto Sans SC, Arial, Helvetica, sans-serif"
EOF
  else
    cat <<EOF > "$style_tmp"
${use_mono:+skinparam monochrome true}
skinparam shadowing false
skinparam roundCorner 20
skinparam dpi ${dpi}
scale ${scale}
skinparam defaultFontSize 14
skinparam defaultFontName "Arial, Helvetica, sans-serif"
skinparam padding 8
skinparam ArrowThickness 2
skinparam BorderThickness 2
skinparam svgDimensionStyle false
skinparam svgLinkTarget _blank
EOF
  fi

  # Insert the style block after the (single) @start line, whatever its type.
  strip_style "$input" | sed "/^[[:space:]]*@start[a-zA-Z]/r ${style_tmp}" > "$output"
  rm -f "$style_tmp"
}

# ── PNG adaptive rendering ────────────────────────────────────────────────────

# Calculate optimal PNG scale and DPI from SVG viewBox dimensions.
# Target: max(width, height) ≤ PNG_MAX
#
# Relationship: PNG_pixels ≈ SVG_viewBox_at_target_scale
# So: target_scale = PNG_MAX / max(svg_w, svg_h) * SVG_SCALE
#
# If target_scale < 1, keep scale=1 and reduce DPI instead:
#   target_dpi = PNG_MAX * PNG_DPI / (max_dim / SVG_SCALE)
calc_png_params() {
  local svg_file="$1"
  local viewbox max_dim target_scale target_dpi

  # Extract viewBox dimensions from SVG
  viewbox=$(grep -oE 'viewBox="[0-9]+ [0-9]+ [0-9]+ [0-9]+"' "$svg_file" 2>/dev/null | head -1)
  if [[ -z "$viewbox" ]]; then
    # Fallback: can't parse viewBox, use safe defaults
    echo "2 150"
    return
  fi

  local vb_w vb_h
  vb_w=$(echo "$viewbox" | grep -oE '[0-9]+' | sed -n '3p')
  vb_h=$(echo "$viewbox" | grep -oE '[0-9]+' | sed -n '4p')

  if [[ -z "$vb_w" ]] || [[ -z "$vb_h" ]]; then
    echo "2 150"
    return
  fi

  # max dimension from SVG rendered at SVG_SCALE
  if (( vb_w > vb_h )); then
    max_dim=$vb_w
  else
    max_dim=$vb_h
  fi

  # Base dimension (at scale 1) = max_dim / SVG_SCALE
  local base_dim=$(( max_dim / SVG_SCALE ))

  # Calculate scale to fit within PNG_MAX at PNG_DPI
  # PNG size ≈ base_dim × scale × (dpi / 300)
  # At dpi 300: PNG size ≈ base_dim × scale
  # Want: base_dim × scale ≤ PNG_MAX
  # scale ≤ PNG_MAX / base_dim
  target_scale=$(( PNG_MAX / base_dim ))

  if (( target_scale >= SVG_SCALE )); then
    # Diagram is small enough for max scale at full DPI
    echo "${SVG_SCALE} ${PNG_DPI}"
  elif (( target_scale >= 1 )); then
    # Use reduced integer scale at full DPI
    echo "${target_scale} ${PNG_DPI}"
  else
    # Scale 1 still exceeds PNG_MAX at dpi 300
    # Reduce DPI: want base_dim × 1 × (dpi/300) ≤ PNG_MAX
    # dpi ≤ PNG_MAX × 300 / base_dim
    target_dpi=$(( PNG_MAX * PNG_DPI / base_dim ))
    if (( target_dpi < PNG_MIN_DPI )); then
      target_dpi=$PNG_MIN_DPI
    fi
    echo "1 ${target_dpi}"
  fi
}

# Validate PNG output is not blank/corrupted.
# Returns 0 if valid, 1 if likely blank.
validate_png() {
  local png_file="$1"
  local file_size dim_info

  if [[ ! -f "$png_file" ]]; then
    return 1
  fi

  file_size=$(wc -c < "$png_file" | tr -d ' ')

  # Check if file hit 4096 cap AND is suspiciously small (likely blank)
  dim_info=$(file "$png_file" 2>/dev/null)
  if echo "$dim_info" | grep -q "4096 x 4096" && (( file_size < PNG_BLANK_THRESHOLD )); then
    return 1
  fi

  # Also check for very small files that indicate rendering failure
  if (( file_size < 1000 )); then
    return 1
  fi

  return 0
}

# ── Main ─────────────────────────────────────────────────────────────────────

main() {
  local input="${1:-}" output_dir="${2:-.}" prefix="${3:-diagram}"

  if [[ -z "$input" ]] || [[ ! -f "$input" ]]; then
    echo "Usage: render-plantuml.sh <input.puml> [output_dir] [output_prefix]" >&2
    exit 1
  fi

  mkdir -p "$output_dir"

  # ── Select rendering backend (server or local jar) ──
  BACKEND="$(select_backend)"
  if [[ "$BACKEND" == "local" ]]; then
    JAR="$(resolve_jar)"
    log "Backend: local jar (${JAR})"
  else
    log "Backend: server (${PLANTUML_SERVER})"
  fi

  # ── Guard: prevent input/output path collision ──
  # When output_dir/prefix.puml resolves to the same file as input,
  # inject_style's shell redirection would truncate the input before reading.
  # Fix: copy input to a temp file and use that as the source.
  local real_input
  real_input=$(cd "$(dirname "$input")" && pwd)/$(basename "$input")
  local real_output
  real_output=$(cd "$output_dir" && pwd)/${prefix}.puml

  local effective_input="$input"
  local _input_tmp=""
  if [[ "$real_input" == "$real_output" ]]; then
    _input_tmp=$(mktemp "${output_dir}/${prefix}.src.XXXXXX.puml")
    cp "$input" "$_input_tmp"
    effective_input="$_input_tmp"
    log "Input/output path collision detected; using temp copy: ${_input_tmp}"
  fi

  local puml="${output_dir}/${prefix}.puml"
  local svg="${output_dir}/${prefix}.svg"
  local png="${output_dir}/${prefix}.png"
  local png_puml="${output_dir}/${prefix}.png.tmp.puml"

  # ── Step 1: Render SVG (always at max quality) ──
  inject_style "$effective_input" "$puml" "$SVG_SCALE" "$SVG_DPI"
  log "SVG style applied (scale=${SVG_SCALE}, dpi=${SVG_DPI})"

  log "Rendering SVG..."
  if ! render_diagram "$puml" "$svg" svg; then
    warn "SVG rendering failed"
    rm -f "$png_puml"
    exit 1
  fi

  # ── Step 2: Calculate optimal PNG parameters ──
  local png_params png_scale png_dpi
  png_params=$(calc_png_params "$svg")
  png_scale=$(echo "$png_params" | cut -d' ' -f1)
  png_dpi=$(echo "$png_params" | cut -d' ' -f2)
  log "PNG adaptive params: scale=${png_scale}, dpi=${png_dpi} (target ≤ ${PNG_MAX}px)"

  # ── Step 3: Render PNG with adaptive parameters ──
  inject_style "$effective_input" "$png_puml" "$png_scale" "$png_dpi"

  log "Rendering PNG..."
  if ! render_diagram "$png_puml" "$png" png; then
    warn "PNG rendering failed"
    rm -f "$png_puml"
    exit 1
  fi

  # ── Step 4: Validate PNG output ──
  if ! validate_png "$png"; then
    warn "PNG output appears blank or corrupted (file too small for 4096×4096)"
    warn "Retrying with reduced parameters..."

    # Fallback: scale 1, dpi 150
    local fallback_scale=1
    local fallback_dpi=150
    inject_style "$effective_input" "$png_puml" "$fallback_scale" "$fallback_dpi"
    log "PNG fallback: scale=${fallback_scale}, dpi=${fallback_dpi}"

    render_diagram "$png_puml" "$png" png || true

    if ! validate_png "$png"; then
      warn "PNG fallback also produced invalid output. PNG may be incomplete."
    fi
  fi

  rm -f "$png_puml"
  [[ -n "$_input_tmp" ]] && rm -f "$_input_tmp"

  # ── Step 5: Report results ──
  local svg_vb png_dim png_size
  svg_vb=$(grep -oE 'viewBox="[^"]*"' "$svg" 2>/dev/null | head -1 || echo "unknown")
  png_dim=$(file "$png" 2>/dev/null | grep -oE '[0-9]+ x [0-9]+' | head -1 || echo "unknown")
  png_size=$(wc -c < "$png" 2>/dev/null | tr -d ' ' || echo "0")

  echo "=== Rendering Complete ==="
  echo "Source: ${puml}"
  echo "SVG:    ${svg} (${svg_vb})"
  echo "PNG:    ${png} (${png_dim}, ${png_size} bytes)"
  echo "Config: SVG=scale${SVG_SCALE}/dpi${SVG_DPI} | PNG=scale${png_scale}/dpi${png_dpi} (target ≤${PNG_MAX}px)"

  # Warn if PNG hit the cap
  if echo "$png_dim" | grep -q "4096"; then
    warn "PNG hit server hard cap (4096). Content may be clipped. Prefer SVG for this diagram."
  fi
}

main "$@"
