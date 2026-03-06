import * as api from '@actual-app/api';

function fail(message) {
  process.stderr.write(String(message));
  process.exit(1);
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
