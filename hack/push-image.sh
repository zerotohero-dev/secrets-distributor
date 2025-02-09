#!/usr/bin/env bash

docker tag localhost:32000/secrets-distributor:v0.1.0 \
  localhost:32000/secrets-distributor:v0.1.0

docker push localhost:32000/secrets-distributor:v0.1.0