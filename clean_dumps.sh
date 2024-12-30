#!/bin/bash

# Configuration
SERVER1_USER="MATS-VM-01_admin"
SERVER1_IP="172.201.121.48"
SSH_KEY="/root/.ssh/Olex.pem"
REMOTE_DUMP_PATH="/var/lib/mongodb/download/dump"
FETCH_SCRIPT_PATH="/home/MATS-VM-01_admin/fetch_latest_dump.sh"

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

# Execute fetch_latest_dump.sh script
echo "Executing fetch_latest_dump.sh..."
if ! sudo bash ${FETCH_SCRIPT_PATH}; then
  echo "Error occurred while executing fetch_latest_dump.sh."
  exit 1
fi
echo "Successfully fetched the latest dumps."

# Drop specified databases
echo "Dropping old databases..."
mongo --eval 'db.getSiblingDB("mats-diagnostic").dropDatabase()'
mongo --eval 'db.getSiblingDB("mats-payment").dropDatabase()'
mongo --eval 'db.getSiblingDB("mats-training-plan").dropDatabase()'
mongo --eval 'db.getSiblingDB("mats-user").dropDatabase()'

# Restore latest dump
echo "Restoring the latest dump..."
mongorestore --gzip /var/lib/mongodb/download/dump

EOF
