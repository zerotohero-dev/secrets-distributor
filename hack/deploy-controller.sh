#!/usr/bin/env bash

KUBECONFIG=$(microk8s config)
export KUBECONFIG=KUBECONFIG

make deploy IMG=localhost:32000/secrets-distributor:v0.1.0