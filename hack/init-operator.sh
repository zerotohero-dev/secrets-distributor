#!/usr/bin/env bash

# This script should run in a "mostly" empty repository to initialize the
# boilerplate code of the Operator.

# See https://sdk.operatorframework.io/docs/installation/
# for installation instructions for operator-sdk

operator-sdk init --domain secrets-distributor.z2h.dev \
  --repo github.com/zerotohero-dev/secrets-distributor
