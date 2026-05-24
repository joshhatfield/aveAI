#!/bin/bash
# aveAI OpenCode Plugin Installer

set -e

REPO="joshhatfield/aveAI"
BRANCH="${AV_BRANCH:-master}"
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
  echo "  curl -fsSL $PLUGIN_URL | bash          # Interactive install"
  echo "  curl -fsSL $PLUGIN_URL | bash -s -- 1  # Local install (non-interactive)"
  echo "  curl -fsSL $PLUGIN_URL | bash -s -- 2  # Global install (non-interactive)"
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
    exit 1
  fi

  # Create .opencode/plugins if they don't exist
  if [ ! -d ".opencode" ]; then
    log "Creating .opencode directory..."
    mkdir -p ".opencode/plugins"
  elif [ ! -d ".opencode/plugins" ]; then
    log "Creating .opencode/plugins directory..."
    mkdir -p ".opencode/plugins"
  fi

  local_plugin_dir=".opencode/plugins"

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

# Non-interactive mode: accept option as argument
if [[ $# -gt 0 ]]; then
  case $1 in
    1)
      install_local
      exit 0
      ;;
    2)
      install_global
      exit 0
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "Invalid option: $1"
      echo "Use 1 for local, 2 for global"
      exit 1
      ;;
  esac
fi

# Interactive mode
echo "aveAI OpenCode Plugin Installer"
echo "================================"
echo ""
echo "1) Local install  (current project: $(basename "$(pwd)"))"
echo "2) Global install (all projects)"
echo ""

read -p "Choose an option [1]: " choice
choice="${choice:-1}"

case $choice in
  1)
    install_local
    ;;
  2)
    install_global
    ;;
  *)
    echo "Invalid option: $choice"
    exit 1
    ;;
esac