#!/bin/bash

# Configuration
SERVER1_USER="MATS-VM-01_admin"
SERVER1_IP="172.201.121.48"
SSH_KEY="/root/.ssh/Olex.pem"
REMOTE_DUMP_PATH="/var/lib/mongodb/download/dump"

echo "Deleting old dumps on server 1..."

# Connect to server 1 and remove old dumps
# shellcheck disable=SC2087
ssh -i "$SSH_KEY" "${SERVER1_USER}@${SERVER1_IP}" <<EOF
echo "Connected to server 1. Removing old dumps..."
if ! sudo rm -rf ${REMOTE_DUMP_PATH}/*; then
  echo "Error occurred while deleting old dumps."
  exit 1
fi
echo "Old dumps successfully deleted."
EOF
