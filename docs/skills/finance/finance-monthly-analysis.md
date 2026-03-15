# Finance Skill: Monthly Analysis

Use this skill layer with the monthly variance workflow to turn output into practical household decisions.

- Command workflow reference: [`../../workflows/finance-monthly-analysis.md`](../../workflows/finance-monthly-analysis.md)
- Mental model reference: [`./mental-models.md`](./mental-models.md)

## Inputs

- Monthly variance output (`raw`, `normalized`, `quality`)
- Category-group deltas vs prior months
- Known recurring bills and planned irregular expenses

## Mental-model checks to apply

1. **Margin of safety**
   - Check emergency buffer coverage (months of essential outflows).
   - If coverage drops below target, prioritize restoring buffer.

2. **Circle of competence**
   - Separate explainable variances (known bills/seasonality) from unknown variances.
   - Lower confidence when unexplained deltas are material.

3. **Inversion**
   - Ask: "What would make next month fail?"
   - Test one adverse scenario (income dip or unavoidable expense spike) and flag pre-emptive cuts.

4. **Long-term compounding mindset**
   - Emphasize repeatable category fixes over one-time austerity.
   - Compare 3–6 month trend direction before judging progress.

5. **Opportunity cost framing**
   - For each proposed spend increase, state what goal is delayed (buffer, debt payoff, or sinking fund).

## Output style

Return concise sections:

1. **What changed** (top variances)
2. **What matters** (risk level + confidence)
3. **What to do this month** (3 concrete actions)
4. **Trade-offs** (explicit opportunity-cost statements)

## Example recommendation phrasing

"Groceries and transport were the main variance drivers. Because your safety buffer is below 2 months, prioritize stabilizing those two categories and postpone non-urgent discretionary upgrades until coverage recovers."