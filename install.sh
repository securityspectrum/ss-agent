#!/bin/bash

# Variables
AGENT_URL="https://example.com/agent"
CONFIG_URL="https://example.com/config"
INSTALL_DIR="/opt/siem-agent"
CONFIG_FILE="$INSTALL_DIR/config.json"

# Create installation directory
sudo mkdir -p $INSTALL_DIR

# Download the agent binary
sudo curl -o $INSTALL_DIR/agent $AGENT_URL
sudo chmod +x $INSTALL_DIR/agent

# Download the configuration file
curl -o $CONFIG_FILE $CONFIG_URL

# Register the agent
sudo $INSTALL_DIR/agent register

echo "SIEM agent installed and started."
