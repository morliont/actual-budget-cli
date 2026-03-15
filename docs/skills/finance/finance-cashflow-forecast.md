# Finance Skill: Cashflow Forecast

Use this skill layer when building short-term household cashflow forecasts (typically 1–3 months).

- Mental model reference: [`./mental-models.md`](./mental-models.md)

## Inputs

- Current cash position and upcoming known inflows/outflows
- Recurring bills, debt minimums, and scheduled irregular expenses
- Recent category spend run-rate

## Mental-model checks to apply

1. **Margin of safety**
   - Forecast minimum end-of-month cash and compare with required buffer floor.
   - If forecasted cash breaches floor, mark "protective mode".

2. **Circle of competence**
   - Use conservative assumptions only from known data.
   - Label uncertain items (variable utilities, unconfirmed reimbursements) explicitly.

3. **Inversion**
   - Run at least one downside scenario: delayed income or +10–15% variable bills.
   - If downside goes negative, propose cuts/timing changes before month start.

4. **Long-term compounding mindset**
   - Prefer stable automation (fixed transfers, sinking funds, debt autopay) that improves each forecast cycle.
   - Measure forecast quality over time and reduce repeated misses.

5. **Opportunity cost framing**
   - When adding planned spending, show impact on runway and goal timelines.

## Output style

1. **Base forecast** (expected position)
2. **Downside check** (stress result)
3. **Actions now** (sequence: protect essentials → stabilize volatility → fund priorities)
4. **Trade-off callouts** (what each optional spend displaces)

## Example recommendation phrasing

"Base case stays positive, but the downside case goes negative in week 4. Move one discretionary purchase to next month and keep the automatic buffer transfer to preserve runway."