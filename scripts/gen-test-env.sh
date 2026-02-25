#!/usr/bin/env bash
# scripts/gen-test-env.sh
# Generates ephemeral secrets for the Lago test stack.
# Source this file: . scripts/gen-test-env.sh

set -euo pipefail

export LAGO_RSA_PRIVATE_KEY=$(openssl genrsa 2048 2>/dev/null | base64 -w0 2>/dev/null || openssl genrsa 2048 2>/dev/null | base64)
export LAGO_ORG_API_KEY="lago-test-$(openssl rand -hex 16)"
export SECRET_KEY_BASE=$(openssl rand -hex 64)
export LAGO_ENCRYPTION_PRIMARY_KEY=$(openssl rand -hex 16)
export LAGO_ENCRYPTION_DETERMINISTIC_KEY=$(openssl rand -hex 16)
export LAGO_ENCRYPTION_KEY_DERIVATION_SALT=$(openssl rand -hex 16)
export POSTGRES_PASSWORD=$(openssl rand -hex 16)
export LAGO_ORG_USER_PASSWORD=$(openssl rand -hex 12)
