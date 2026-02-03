#!/bin/bash
# MyOps Agent Installation Script

set -e

VERSION="1.0.0"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/myops-agent"
SERVICE_FILE="/etc/systemd/system/myops-agent.service"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        OS_VERSION=$VERSION_ID
    else
        echo_error "Cannot detect OS type"
        exit 1
    fi

    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            AGENT_ARCH="amd64"
            ;;
        aarch64)
            AGENT_ARCH="arm64"
            ;;
        *)
            echo_error "Unsupported architecture: $ARCH"
            exit 1
            ;;
    esac

    echo_info "Detected OS: $OS $OS_VERSION ($AGENT_ARCH)"
}

# Check if running as root
check_root() {
    if [ "$EUID" -ne 0 ]; then
        echo_error "This script must be run as root"
        exit 1
    fi
}

# Get server endpoint from user
get_config() {
    if [ -z "$MYOPS_SERVER_ENDPOINT" ]; then
        read -p "Enter MyOps server endpoint (e.g., https://myops.example.com): " SERVER_ENDPOINT
        MYOPS_SERVER_ENDPOINT=$SERVER_ENDPOINT
    fi

    if [ -z "$MYOPS_AGENT_TOKEN" ]; then
        read -p "Enter agent token: " AGENT_TOKEN
        MYOPS_AGENT_TOKEN=$AGENT_TOKEN
    fi

    if [ -z "$MYOPS_SERVER_ENDPOINT" ] || [ -z "$MYOPS_AGENT_TOKEN" ]; then
        echo_error "Server endpoint and token are required"
        exit 1
    fi
}

# Create config file
create_config() {
    echo_info "Creating configuration file..."

    mkdir -p "$CONFIG_DIR"

    cat > "$CONFIG_DIR/config.yaml" <<EOF
# MyOps Agent Configuration
server:
  endpoint: "$MYOPS_SERVER_ENDPOINT"
  token: "$MYOPS_AGENT_TOKEN"
  insecure: false

report:
  interval: 60

collector:
  collect_processes: false
  collect_network: true
EOF

    chmod 600 "$CONFIG_DIR/config.yaml"
    echo_info "Configuration saved to $CONFIG_DIR/config.yaml"
}

# Install systemd service
install_service() {
    if command -v systemctl &> /dev/null; then
        echo_info "Installing systemd service..."

        cat > "$SERVICE_FILE" <<EOF
[Unit]
Description=MyOps Agent
After=network.target

[Service]
Type=simple
User=root
ExecStart=$INSTALL_DIR/myops-agent
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

        systemctl daemon-reload
        systemctl enable myops-agent
        echo_info "Systemd service installed and enabled"
    else
        echo_warn "systemd not found, skipping service installation"
    fi
}

# Start the service
start_service() {
    if command -v systemctl &> /dev/null; then
        echo_info "Starting myops-agent service..."
        systemctl start myops-agent
        systemctl status myops-agent --no-pager
    fi
}

# Main installation
main() {
    echo_info "MyOps Agent Installer v$VERSION"
    echo ""

    check_root
    detect_os
    get_config
    create_config
    install_service
    start_service

    echo ""
    echo_info "Installation completed successfully!"
    echo_info "To view logs: journalctl -u myops-agent -f"
    echo_info "To restart: systemctl restart myops-agent"
    echo_info "To stop: systemctl stop myops-agent"
}

main "$@"
