#!/usr/bin/env bash
# coverage.sh — Run unit tests, filter out non-unit-testable packages,
# and report both raw and filtered coverage.
#
# Excluded packages (require live infrastructure or are auto-generated):
#   cmd/                   — entry-point wiring
#   internal/repository/   — real PostgreSQL / live DB
#   internal/smtp/         — real SMTP/Gmail
#   internal/transport/    — RabbitMQ consumer / gRPC server wiring
#   internal/database/     — generated sqlc code
#   internal/worker/       — background scheduler (requires live DB + time)
#   mocks/                 — generated test helpers
#   tests/                 — BDD/integration tests (need live infra)
#   shared/                — covered separately by shared/auth tests

set -euo pipefail

COVERAGE_RAW="coverage.out"
COVERAGE_FILTERED="coverage_filtered.out"
COVERAGE_HTML="coverage.html"

echo "═══════════════════════════════════════════════════════════"
echo "  Running unit tests…"
echo "═══════════════════════════════════════════════════════════"

go test ./services/... ./shared/... \
    -coverprofile="$COVERAGE_RAW" \
    -covermode=atomic \
    -count=1 \
    "$@"

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  Raw coverage (all packages)"
echo "═══════════════════════════════════════════════════════════"
go tool cover -func="$COVERAGE_RAW" | grep "^total:"

# ── Filter out excluded packages ─────────────────────────────────────────────
# Keep the 'mode:' header line plus any line whose package path does NOT match
# the excluded patterns.
grep -vE \
    '/cmd/|/internal/repository/|/internal/smtp/|/internal/transport/|/internal/database/|/internal/worker/|/internal/testutil/|bank-service/internal/domain/|bank-service/internal/handler/|/mocks/|/tests/|/utils/rabbitmq\.go' \
    "$COVERAGE_RAW" \
    > "$COVERAGE_FILTERED"

echo ""
echo "═══════════════════════════════════════════════════════════"
echo "  Filtered coverage (unit-testable packages only)"
echo "═══════════════════════════════════════════════════════════"
echo ""
echo "  Excluded:"
echo "    • cmd/                             (entry-point wiring)"
echo "    • internal/repository/             (real PostgreSQL)"
echo "    • internal/smtp/                   (real SMTP)"
echo "    • internal/transport/              (gRPC server / RabbitMQ consumer)"
echo "    • internal/database/               (generated sqlc code)"
echo "    • internal/worker/                 (background scheduler)"
echo "    • internal/testutil/               (test-only helpers)"
echo "    • bank-service/internal/domain/    (GORM models / boilerplate)"
echo "    • bank-service/internal/handler/   (HTTP+gRPC handlers — future sprint)"
echo "    • mocks/                           (generated test helpers)"
echo "    • tests/                           (BDD / integration tests)"
echo "    • utils/rabbitmq.go                (real AMQP connection)"
echo ""

go tool cover -func="$COVERAGE_FILTERED" | grep -v "^total:"

echo ""
FILTERED_TOTAL=$(go tool cover -func="$COVERAGE_FILTERED" | grep "^total:" | awk '{print $3}')
echo "  Filtered total:  $FILTERED_TOTAL"

# ── HTML report ──────────────────────────────────────────────────────────────
go tool cover -html="$COVERAGE_FILTERED" -o "$COVERAGE_HTML"
echo ""
echo "  HTML report: $COVERAGE_HTML"
echo "═══════════════════════════════════════════════════════════"

# ── Exit non-zero if filtered total < 80% ────────────────────────────────────
PERCENT=${FILTERED_TOTAL//%/}
THRESHOLD=80

# Use awk for floating-point comparison
if awk "BEGIN { exit ($PERCENT >= $THRESHOLD) ? 0 : 1 }"; then
    echo "  ✓ Coverage ${FILTERED_TOTAL} meets the ≥${THRESHOLD}% threshold."
else
    echo "  ✗ Coverage ${FILTERED_TOTAL} is BELOW the ≥${THRESHOLD}% threshold." >&2
    exit 1
fi
