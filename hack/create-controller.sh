#!/usr/bin/env bash

# Create the controller
operator-sdk create api --group "" --version v1 --kind Secret --controller