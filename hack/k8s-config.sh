#!/usr/bin/env bash

microk8s config > ~/.kube/microk8s-config

echo "Run 'export KUBECONFIG=~/.kube/microk8s-config'"

echo ""