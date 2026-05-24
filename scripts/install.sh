#!/bin/bash
# aveAI OpenCode Plugin Installer
# Usage: curl -fsSL https://raw.githubusercontent.com/joshhatfield/aveAI/main/scripts/install.sh | bash
# Or with global flag: curl -fsSL ... | bash -s -- --global

set -e

REPO="joshhatfield/aveAI"
BRANCH="${AV_BRANCH:-main}"
RAW_BASE="https://raw.githubusercontent.com/${REPO}/${BRANCH}"
PLUGIN_FILE="ave.ts"
PLUGIN_URL="${RAW_BASE}/.opencode/plugins/${PLUGIN_FILE}"

# Global install directory
GLOBAL_DIR="${HOME}/.config/opencode/plugins"
GLOBAL_PLUGIN_DIR="${GLOBAL_DIR}/ave"

usage() {
  echo "aveAI OpenCode Plugin Installer"
  echo ""
  echo "Usage:"
  echo "  curl -fsSL $PLUGIN_URL | bash          # Local install (current dir)"
  echo "  curl -fsSL $PLUGIN_URL | bash -s -- --global  # Global install"
  echo ""
  echo "The plugin will be installed to:"
  echo "  Local:  <project>/.opencode/plugins/ave.ts"
  echo "  Global: ~/.config/opencode/plugins/ave/ave.ts"
}

log() {
  echo "[ave-install] $1"
}

install_local() {
  local_project_dir="$(pwd)"

  # Check if we're in a git repo
  if [ ! -d ".git" ]; then
    log "Error: Not a git repository. Run from a project directory."
    log "Or use --global for global installation."
    exit 1
  fi

  # Check if .opencode directory exists
  if [ ! -d ".opencode" ]; then
    log "Error: No .opencode directory found. Run 'opencode init' first."
    exit 1
  fi

  local_plugin_dir=".opencode/plugins"
  mkdir -p "$local_plugin_dir"

  log "Installing ave plugin to ${local_project_dir}/${local_plugin_dir}/..."
  if curl -fsSL "$PLUGIN_URL" -o "${local_plugin_dir}/${PLUGIN_FILE}"; then
    log "✓ Installed ${local_plugin_dir}/${PLUGIN_FILE}"
  else
    log "Error: Failed to download plugin from $PLUGIN_URL"
    exit 1
  fi

  # Verify it's valid TypeScript (basic check)
  if grep -q "export default" "${local_plugin_dir}/${PLUGIN_FILE}"; then
    log "✓ Plugin file verified"
  else
    log "Warning: Plugin file may not be valid"
  fi

  log ""
  log "Local installation complete!"
  log "Restart OpenCode or reload to use the ave tool."
}

install_global() {
  log "Installing ave plugin globally to ${GLOBAL_PLUGIN_DIR}/..."

  mkdir -p "$GLOBAL_PLUGIN_DIR"

  if curl -fsSL "$PLUGIN_URL" -o "${GLOBAL_PLUGIN_DIR}/${PLUGIN_FILE}"; then
    log "✓ Installed ${GLOBAL_PLUGIN_DIR}/${PLUGIN_FILE}"
  else
    log "Error: Failed to download plugin from $PLUGIN_URL"
    exit 1
  fi

  if grep -q "export default" "${GLOBAL_PLUGIN_DIR}/${PLUGIN_FILE}"; then
    log "✓ Plugin file verified"
  else
    log "Warning: Plugin file may not be valid"
  fi

  log ""
  log "Global installation complete!"
  log "Restart OpenCode or reload to use the ave tool."
}

# Parse arguments
GLOBAL_FLAG=false
while [[ $# -gt 0 ]]; do
  case $1 in
    --global)
      GLOBAL_FLAG=true
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      shift
      ;;
  esac
done

# Detect mode and install
if [ "$GLOBAL_FLAG" = true ]; then
  install_global
elif [ -d ".git" ] && [ -d ".opencode/plugins" ]; then
  install_local
elif [ -d "$GLOBAL_DIR" ]; then
  install_global
else
  log "Error: No install target found."
  log ""
  log "Options:"
  log "  1. Run from a project with .opencode/plugins/ directory"
  log "  2. Run with --global flag"
  log "  3. Run 'opencode init' first to create .opencode/"
  exit 1
fi