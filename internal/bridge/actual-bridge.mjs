import * as api from '@actual-app/api';
import fs from 'node:fs';
import path from 'node:path';
import { createRequire } from 'node:module';
import { fileURLToPath } from 'node:url';

const require = createRequire(import.meta.url);
const bridgeScriptDir = path.dirname(fileURLToPath(import.meta.url));

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

function isMigrationMismatch(err) {
  return String(err?.message || err).includes('out-of-sync-migrations') ||
    String(err?.message || err).includes('Database is out of sync with migrations');
}

function isSqliteFallbackOp(op) {
  return ['accounts-list', 'categories-list', 'transactions-list', 'budgets-categories', 'budgets-summary'].includes(op);
}

function monthToKey(month) {
  return String(month || '').replace('-', '');
}

function dateToInt(date) {
  return Number(String(date || '').replaceAll('-', ''));
}

function intToDate(value) {
  const s = String(value || '');
  if (s.length !== 8) return s;
  return `${s.slice(0, 4)}-${s.slice(4, 6)}-${s.slice(6, 8)}`;
}

function readInt(value, fallback = 0) {
  const n = Number(value);
  return Number.isFinite(n) ? n : fallback;
}

function findBudgetDir(dataDir, budgetId) {
  const entries = fs.readdirSync(dataDir, { withFileTypes: true });
  const candidates = [];

  for (const entry of entries) {
    if (!entry.isDirectory()) continue;
    const dir = path.join(dataDir, entry.name);
    const dbPath = path.join(dir, 'db.sqlite');
    if (!fs.existsSync(dbPath)) continue;

    let score = 0;
    const metadataPath = path.join(dir, 'metadata.json');
    if (fs.existsSync(metadataPath)) {
      try {
        const metadata = JSON.parse(fs.readFileSync(metadataPath, 'utf8'));
        if (metadata?.groupId === budgetId || metadata?.cloudFileId === budgetId || metadata?.id === budgetId) {
          score += 10;
        }
      } catch {
        // Ignore invalid metadata and keep the directory as a lower-confidence candidate.
      }
    }

    const stat = fs.statSync(dbPath);
    candidates.push({ dir, score, mtimeMs: stat.mtimeMs });
  }

  candidates.sort((a, b) => (b.score - a.score) || (b.mtimeMs - a.mtimeMs));
  if (!candidates.length) {
    throw new Error(`no local Actual SQLite budget cache found under ${dataDir}`);
  }
  return candidates[0].dir;
}

function openLocalDatabases(cfg) {
  const Database = require('better-sqlite3');
  const budgetDir = findBudgetDir(cfg.dataDir, cfg.budgetId);
  const db = new Database(path.join(budgetDir, 'db.sqlite'), { readonly: true, fileMustExist: true });
  const cachePath = path.join(budgetDir, 'cache.sqlite');
  const cache = fs.existsSync(cachePath)
    ? new Database(cachePath, { readonly: true, fileMustExist: true })
    : null;

  return {
    db,
    cache,
    close() {
      db.close();
      if (cache) cache.close();
    },
  };
}

function maxAvailableMigrationId() {
  const roots = [
    path.join(bridgeScriptDir, 'node_modules', '@actual-app', 'api', 'dist', 'migrations'),
    path.join(bridgeScriptDir, 'node_modules', '@actual-app', 'core', 'migrations'),
    path.join(process.cwd(), 'node_modules', '@actual-app', 'api', 'dist', 'migrations'),
    path.join(process.cwd(), 'node_modules', '@actual-app', 'core', 'migrations'),
  ];
  let max = 0;
  for (const root of roots) {
    if (!fs.existsSync(root)) continue;
    for (const name of fs.readdirSync(root)) {
      const match = name.match(/^(\d+)/);
      if (!match) continue;
      max = Math.max(max, Number(match[1]));
    }
  }
  return max;
}

function localDatabaseOutrunsApi(cfg) {
  let local;
  try {
    local = openLocalDatabases(cfg);
    const applied = local.db.prepare('select max(id) as max_id from __migrations__').get()?.max_id || 0;
    const available = maxAvailableMigrationId();
    return available > 0 && Number(applied) > available;
  } catch {
    return false;
  } finally {
    if (local) local.close();
  }
}

function cacheValue(cache, key, fallback = 0) {
  if (!cache) return fallback;
  const row = cache.prepare('select value from kvcache where key = ?').get(key);
  return row ? readInt(row.value, fallback) : fallback;
}

function cacheBool(cache, key) {
  if (!cache) return null;
  const row = cache.prepare('select value from kvcache where key = ?').get(key);
  if (!row) return null;
  if (row.value === 'true') return true;
  if (row.value === 'false') return false;
  return null;
}

function localCategories(db) {
  return db.prepare(`
    select
      c.id,
      c.name,
      c.cat_group as group_id,
      g.name as group_name,
      c.hidden,
      null as archived,
      c.sort_order,
      g.sort_order as group_sort_order,
      c.is_income,
      g.is_income as group_is_income
    from categories c
    left join category_groups g on g.id = c.cat_group and g.tombstone = 0
    where c.tombstone = 0
    order by g.sort_order, c.sort_order, c.name
  `).all().map((row) => ({
    ...row,
    hidden: Boolean(row.hidden),
    is_income: Boolean(row.is_income),
    group_is_income: Boolean(row.group_is_income),
  }));
}

function sqliteFallback(cfg, op, args) {
  const local = openLocalDatabases(cfg);
  try {
    const { db, cache } = local;

    if (op === 'accounts-list') {
      const accounts = db.prepare(`
        select id, name, type, offbudget, closed, balance_current
        from accounts
        where tombstone = 0
        order by sort_order, name
      `).all().map((row) => ({
        ...row,
        offbudget: Boolean(row.offbudget),
        closed: Boolean(row.closed),
      }));
      return { accounts, source: 'local-sqlite-fallback' };
    }

    if (op === 'categories-list') {
      return { categories: localCategories(db), source: 'local-sqlite-fallback' };
    }

    if (op === 'transactions-list') {
      const from = dateToInt(args.from || '1900-01-01');
      const to = dateToInt(args.to || '2999-12-31');
      const params = [from, to];
      let accountFilter = '';
      if (args.accountId) {
        accountFilter = 'and t.acct = ?';
        params.push(args.accountId);
      }

      const rows = db.prepare(`
        select
          t.id,
          t.date,
          t.acct as account,
          a.name as account_name,
          t.amount,
          t.notes,
          t.category,
          c.name as category_name,
          g.name as category_group_name,
          p.name as payee_name
        from transactions t
        left join accounts a on a.id = t.acct and a.tombstone = 0
        left join categories c on c.id = t.category and c.tombstone = 0
        left join category_groups g on g.id = c.cat_group and g.tombstone = 0
        left join payees p on p.id = t.description and p.tombstone = 0
        where t.tombstone = 0
          and t.date >= ?
          and t.date <= ?
          ${accountFilter}
        order by t.date desc, t.sort_order desc, t.id
        limit ?
      `).all(...params, args.limit && Number.isFinite(args.limit) ? args.limit : 100).map((row) => {
        const out = {
          id: row.id,
          date: intToDate(row.date),
          account: row.account,
          account_name: row.account_name || null,
          payee_name: row.payee_name || null,
          amount: row.amount,
          notes: row.notes || '',
          category: row.category || null,
        };
        if (args.includeCategoryNames) {
          out.category_name = row.category_name || null;
          out.category_group_name = row.category_group_name || null;
        }
        return out;
      });
      return { transactions: rows, source: 'local-sqlite-fallback' };
    }

    if (op === 'budgets-categories') {
      const keyMonth = monthToKey(args.month);
      const from = Number(`${keyMonth}01`);
      const to = Number(`${keyMonth}31`);
      const categories = localCategories(db).map((category) => {
        const budgeted = cacheValue(cache, `budget${keyMonth}!budget-${category.id}`, 0);
        const spent = cacheValue(cache, `budget${keyMonth}!sum-amount-${category.id}`, 0);
        const remaining = cacheValue(cache, `budget${keyMonth}!leftover-${category.id}`, budgeted + spent);
        const carryover = cacheBool(cache, `budget${keyMonth}!carryover-${category.id}`);

        return {
          month: args.month,
          category_id: category.id,
          category_name: category.name,
          category_group_id: category.group_id || '',
          category_group_name: category.group_name || '',
          budgeted,
          planned: budgeted,
          spent,
          actual: spent,
          received: category.is_income ? spent : null,
          is_income: category.is_income,
          group_is_income: category.group_is_income,
          remaining,
          variance: remaining,
          carryover,
          raw: { source: 'local-sqlite-fallback', date_range: { from, to } },
        };
      });
      return { month: args.month, categories, source: 'local-sqlite-fallback' };
    }

    if (op === 'budgets-summary') {
      const now = new Date();
      const month = `${now.getUTCFullYear()}-${String(now.getUTCMonth() + 1).padStart(2, '0')}`;
      const keyMonth = monthToKey(month);
      const incomeGroup = db.prepare('select id from category_groups where is_income = 1 and tombstone = 0 limit 1').get();
      const income = incomeGroup ? cacheValue(cache, `budget${keyMonth}!group-sum-amount-${incomeGroup.id}`, 0) : 0;
      const budgeted = cacheValue(cache, `budget${keyMonth}!total-budgeted`, 0);
      const spent = cacheValue(cache, `budget${keyMonth}!total-spent`, 0);
      return { month, budget: { income, budgeted, spent, source: 'local-sqlite-fallback' } };
    }

    throw new Error(`local SQLite fallback does not support operation: ${op}`);
  } finally {
    local.close();
  }
}

async function accountsWithBalances() {
  const accounts = await api.getAccounts();

  return Promise.all(accounts.map(async (account) => ({
    ...account,
    balance_current: await api.getAccountBalance(account.id),
  })));
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

  let result;
  if (isSqliteFallbackOp(op) && localDatabaseOutrunsApi(cfg)) {
    result = sqliteFallback(cfg, op, args);
    process.stdout.write(JSON.stringify(result));
    return;
  }

  try {
    result = await withSession(cfg, async () => {
    if (op === 'accounts-list') {
      return { accounts: await accountsWithBalances() };
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
          const isIncome = Boolean(category?.is_income || group?.is_income);
          const budgeted = pickNumber(category, ['budgeted', 'budget', 'budgetedAmount']) ?? 0;
          const spent = pickNumber(category, ['spent', 'activity', 'spentAmount']) ?? 0;
          const received = pickNumber(category, ['received']);
          const effectiveActivity = isIncome ? (received ?? spent) : spent;
          const remaining = pickNumber(category, ['remaining', 'balance', 'available']) ?? (budgeted - effectiveActivity);
          const variance = pickNumber(category, ['variance']) ?? (budgeted - effectiveActivity);
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
            received,
            is_income: category?.is_income ?? null,
            group_is_income: group?.is_income ?? null,
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
  } catch (e) {
    if (!isMigrationMismatch(e) || !isSqliteFallbackOp(op)) {
      throw e;
    }
    result = sqliteFallback(cfg, op, args);
  }

  process.stdout.write(JSON.stringify(result));
}

run().catch((e) => {
  fail(e?.message || e);
});
