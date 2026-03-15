# Workflow: Finance Monthly Variance Analysis

Use this workflow when operators or subagents need a deterministic month-over-month budget variance view.

Decision layer reference: [`../skills/finance/finance-monthly-analysis.md`](../skills/finance/finance-monthly-analysis.md)

## Command

```bash
actual-cli reports monthly-variance --from YYYY-MM --to YYYY-MM [--strict] [--json]
```

Machine-safe mode (recommended for orchestration):

```bash
actual-cli --agent-json --non-interactive reports monthly-variance --from 2026-01 --to 2026-03 --strict
```

## What the command does

1. Pulls `budgets categories` data for each month in the inclusive range.
2. Aggregates per month and per category-group with explicit spend semantics:
   - `budgeted`
   - `net_spent` (native Actual signed value)
   - `outflow_spent` (expense-only magnitude from negative spent)
   - `inflow_offsets` (positive offsets/refunds/inflow effects)
   - `net_variance` (`budgeted - net_spent`)
   - `planning_variance` (`budgeted - outflow_spent`, default for planning interpretation)
   - `remaining`
3. Keeps backward-compatible aliases in `raw`:
   - `spent` = `net_spent`
   - `variance` = `net_variance`
4. Emits both:
   - `raw` totals (explicit native + planning variance fields)
   - `normalized` sign-safe summary (`budgetedAbs`, `spentAbs`, signed remaining/net variance + planning variance)
5. Runs explicit reconciliation checks for each month.
6. Adds quality metadata (`confidence`, warning list, check counts).

## Strict mode policy

- `--strict` turns reconciliation mismatches into a hard failure.
- Without `--strict`, command succeeds and reports warnings/failed checks in `quality`.

Strict mode should be enabled for CI automation, scheduled reporting, and handoffs that require trusted totals.

## Stop conditions

Stop and do not continue downstream actions when any of the following is true:

- command exits non-zero in strict mode
- `quality.confidence = low`
- `quality.failedMonthCount > 0`
- required month range is invalid (`from > to`)

## Interpretation guidance

- Use `raw.net_spent` for accounting-native reconciliation with Actual semantics.
- Use `raw.outflow_spent` and `raw.planning_variance` for planning narratives and budget-control decisions.
- `raw.inflow_offsets` explains how much refunds/inflows reduced net spend.
- Use `normalized` when comparing magnitude trends across months without sign confusion.
- Confidence levels:
  - `high`: no failed reconciliation checks
  - `medium`: one failed month/check bucket
  - `low`: multiple failed month/check buckets

## Examples

```bash
# Human-readable table
actual-cli reports monthly-variance --from 2026-01 --to 2026-03

# JSON (no envelope)
actual-cli reports monthly-variance --from 2026-01 --to 2026-03 --json

# Deterministic envelope for agents
actual-cli --agent-json --non-interactive reports monthly-variance --from 2026-01 --to 2026-03

# Strict enforcement for automation pipelines
actual-cli --agent-json --non-interactive reports monthly-variance --from 2026-01 --to 2026-03 --strict
```
