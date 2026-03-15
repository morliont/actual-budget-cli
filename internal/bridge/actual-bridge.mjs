import * as api from '@actual-app/api';

// Keep bridge stdout strictly machine-readable JSON.
// Some upstream libs may emit informational logs; route them to stderr.
const forwardToStderr = (...args) => {
  try {
    process.stderr.write(`${args.map((a) => String(a)).join(' ')}\n`);
  } catch {
    // no-op
  }
};
console.log = forwardToStderr;
console.info = forwardToStderr;
console.debug = () => {};

function fail(message) {
  process.stderr.write(String(message));
  process.exit(1);
}

function readNumber(value) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value;
  }
  if (typeof value === 'string' && value.trim() !== '') {
    const n = Number(value);
    if (Number.isFinite(n)) {
      return n;
    }
  }
  return null;
}

function pickNumber(obj, keys) {
  for (const key of keys) {
    const n = readNumber(obj?.[key]);
    if (n !== null) {
      return n;
    }
  }
  return null;
}

function withNull(v) {
  return v === undefined ? null : v;
}

async function readStdin() {
  const chunks = [];
  for await (const chunk of process.stdin) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks).toString('utf8').trim();
}

async function withSession(cfg, fn) {
  await api.init({
    dataDir: cfg.dataDir,
    serverURL: cfg.serverUrl,
    password: cfg.password,
  });

  try {
    await api.downloadBudget(cfg.budgetId, cfg.budgetPassword ? { password: cfg.budgetPassword } : undefined);
    return await fn();
  } finally {
    await api.shutdown();
  }
}

async function categoriesWithGroups() {
  const groups = await api.getCategoryGroups();
  const categories = await api.getCategories();

  const groupMap = new Map();
  for (const group of groups) {
    groupMap.set(group.id, group);
  }

  const rows = [];
  for (const item of categories) {
    if (!item || typeof item !== 'object' || !('group_id' in item)) {
      continue;
    }

    const group = groupMap.get(item.group_id);
    rows.push({
      id: item.id,
      name: item.name,
      group_id: item.group_id,
      group_name: group?.name || '',
      hidden: Boolean(item.hidden),
      archived: withNull(item.archived),
    });
  }

  rows.sort((a, b) => {
    if (a.group_name !== b.group_name) return a.group_name.localeCompare(b.group_name);
    if (a.name !== b.name) return a.name.localeCompare(b.name);
    return String(a.id).localeCompare(String(b.id));
  });

  return rows;
}

async function run() {
  const [, , op] = process.argv;
  if (!op) fail('missing op');

  const rawPayload = await readStdin();
  if (!rawPayload) fail('missing payload on stdin');

  const payload = JSON.parse(rawPayload);
  const cfg = payload.config;
  const args = payload.args || {};

  const result = await withSession(cfg, async () => {
    if (op === 'accounts-list') {
      return { accounts: await api.getAccounts() };
    }

    if (op === 'categories-list') {
      return { categories: await categoriesWithGroups() };
    }

    if (op === 'transactions-list') {
      let transactions = [];
      if (args.accountId) {
        transactions = await api.getTransactions(args.accountId, args.from, args.to);
      } else {
        const accounts = await api.getAccounts();
        for (const account of accounts) {
          const t = await api.getTransactions(account.id, args.from, args.to);
          transactions.push(...t);
        }
      }

      if (args.includeCategoryNames) {
        const categories = await categoriesWithGroups();
        const categoryMap = new Map();
        for (const c of categories) {
          categoryMap.set(c.id, c);
        }

        transactions = transactions.map((tx) => {
          const c = tx.category ? categoryMap.get(tx.category) : null;
          return {
            ...tx,
            category_name: c?.name || null,
            category_group_name: c?.group_name || null,
          };
        });
      }

      transactions.sort((a, b) => (a.date < b.date ? 1 : -1));
      if (args.limit && Number.isFinite(args.limit)) {
        transactions = transactions.slice(0, args.limit);
      }
      return { transactions };
    }

    if (op === 'budgets-summary') {
      const d = new Date();
      const month = `${d.getUTCFullYear()}-${String(d.getUTCMonth() + 1).padStart(2, '0')}`;
      const budget = await api.getBudgetMonth(month);
      return { month, budget };
    }

    if (op === 'budgets-categories') {
      const budget = await api.getBudgetMonth(args.month);
      const categories = [];
      const groups = Array.isArray(budget?.categoryGroups) ? budget.categoryGroups : [];

      for (const group of groups) {
        const categoryRows = Array.isArray(group?.categories) ? group.categories : [];
        for (const category of categoryRows) {
          const budgeted = pickNumber(category, ['budgeted', 'budget', 'budgetedAmount']) ?? 0;
          const spent = pickNumber(category, ['spent', 'activity', 'spentAmount']) ?? 0;
          const remaining = pickNumber(category, ['remaining', 'balance', 'available']) ?? (budgeted - spent);
          const variance = pickNumber(category, ['variance']) ?? (budgeted - spent);
          const carryover = category?.carryover ?? category?.is_carryover ?? category?.rollover ?? null;

          categories.push({
            month: args.month,
            category_id: category?.id || '',
            category_name: category?.name || '',
            category_group_id: group?.id || category?.group_id || '',
            category_group_name: group?.name || '',
            budgeted,
            planned: budgeted,
            spent,
            actual: spent,
            remaining,
            variance,
            carryover,
            carryover_amount: pickNumber(category, ['carryoverAmount', 'carryover_amount', 'fromLastMonth']),
            raw: category,
          });
        }
      }

      categories.sort((a, b) => {
        if (a.category_group_name !== b.category_group_name) return a.category_group_name.localeCompare(b.category_group_name);
        if (a.category_name !== b.category_name) return a.category_name.localeCompare(b.category_name);
        return String(a.category_id).localeCompare(String(b.category_id));
      });

      return { month: args.month, categories };
    }

    if (op === 'auth-check') {
      const budgets = await api.getBudgets();
      return { ok: true, budgets };
    }

    fail(`unknown operation: ${op}`);
  });

  process.stdout.write(JSON.stringify(result));
}

run().catch((e) => {
  fail(e?.message || e);
});
