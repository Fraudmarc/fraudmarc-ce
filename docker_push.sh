#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push fraudmarc/fraudmarc-ce
docker push fraudmarc/fraudmarc-ce-install
