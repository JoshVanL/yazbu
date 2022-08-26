#!/usr/bin/env bash

if [[ -z "$YAZBU_E2E" ]]; then
    echo "Looks like you are trying to manually run this script. That's a bad idea. Try 'nix-build ./hack/test.nix' instead."
    exit 1
fi

set -eu

yazbu list --config /etc/yazbu/config.yaml
