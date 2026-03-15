# Finance Skill: Anomaly Detection

Use this skill layer when identifying unusual household spending, income changes, or category spikes.

- Mental model reference: [`./mental-models.md`](./mental-models.md)

## Inputs

- Recent transactions and category aggregates
- Recurring-merchant patterns and expected bill ranges
- Prior-month and trailing-average baselines

## Mental-model checks to apply

1. **Margin of safety**
   - Prioritize anomalies that threaten essentials coverage or cash buffer first.
   - Severity should increase when anomaly materially reduces runway.

2. **Circle of competence**
   - Only classify as high-confidence anomaly when baseline behavior is known.
   - Mark unknown merchants/categories as "needs review" rather than overconfident labels.

3. **Inversion**
   - Ask: "If ignored, what failure could this create in 30 days?"
   - Escalate anomalies likely to cause overdraft risk, missed bills, or debt rollover.

4. **Long-term compounding mindset**
   - Track repeated small leaks (subscriptions, fees, convenience spending).
   - Recommend durable fixes (cancel, cap, automate checks), not one-off warnings.

5. **Opportunity cost framing**
   - Quantify annualized impact of recurring anomalies and what goals they displace.

## Output style

1. **Top anomalies** (with confidence)
2. **Risk-ranked impact** (buffer/runway effect)
3. **Recommended fix** (single next action per anomaly)
4. **Trade-off note** (what recurring leak costs over time)

## Example recommendation phrasing

"This new recurring charge looks small monthly, but annualized it competes directly with your emergency-fund target. If it’s non-essential, cancel now and redirect that amount to your buffer."