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
mongosh --eval 'db.getSiblingDB("mats-diagnostic").dropDatabase()'
mongosh --eval 'db.getSiblingDB("mats-payment").dropDatabase()'
mongosh --eval 'db.getSiblingDB("mats-training-plan").dropDatabase()'
mongosh --eval 'db.getSiblingDB("mats-user").dropDatabase()'

# Restore latest dump
echo "Restoring the latest dump..."
timeout 10m mongorestore --gzip /var/lib/mongodb/download/dump

# MongoDB processing scripts after restore
echo "Running MongoDB processing scripts..."

# Script 1: Copy data between mats-training-plan and mats-payment
timeout 10m mongosh --eval '
const srcDB = db.getSiblingDB("mats-training-plan");
const destDB = db.getSiblingDB("mats-payment");

// Assuming you want to copy specific collections, for example, "trainingPlans"
const documents = srcDB.trainingPlan.find().toArray();
if (documents.length > 0) {
    destDB.tempTrainingPlan.insertMany(documents);
}
'

# Script 2: Copy stripePayment collection from mats-payment to mats-user
timeout 10m mongosh --eval '
const srcDB = db.getSiblingDB("mats-payment");
const destDB = db.getSiblingDB("mats-user");

// Get all documents from the stripePayment collection
const documents = srcDB.stripePayment.find().toArray();
if (documents.length > 0) {
    destDB.tempStripePayment.insertMany(documents);
}
'

# Script 3: Copy userSubscription collection from mats-payment to mats-user
timeout 10m mongosh --eval '
const srcDB = db.getSiblingDB("mats-payment");
const destDB = db.getSiblingDB("mats-user");

// Get all documents from the userSubscription collection
const documents = srcDB.userSubscription.find().toArray();
if (documents.length > 0) {
    destDB.tempUserSubscription.insertMany(documents);
}
'

# Script 4: Copy consumableProduct collection from mats-payment to mats-user
timeout 10m mongosh --eval '
const srcDB = db.getSiblingDB("mats-payment");
const destDB = db.getSiblingDB("mats-user");

// Get all documents from the consumableProduct collection
const documents = srcDB.consumableProduct.find().toArray();
if (documents.length > 0) {
    destDB.tempConsumableProduct.insertMany(documents);
}
'
echo "Copying executed successfully."

EOF

echo "Script executed successfully."