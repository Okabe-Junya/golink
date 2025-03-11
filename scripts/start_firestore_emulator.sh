#!/bin/bash

echo "Starting Firestore emulator on port 8081..."
gcloud emulators firestore start --host-port=0.0.0.0:8081
