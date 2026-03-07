#!/usr/bin/env node

import fs from 'node:fs';

function argValue(flag) {
  const i = process.argv.indexOf(flag);
  return i >= 0 ? process.argv[i + 1] : undefined;
}

const hasFlag = (flag) => process.argv.includes(flag);

const payloadPath = argValue('--payload');
const url = argValue('--url');
const useApi = hasFlag('--api');
const dryRun = hasFlag('--dry-run');

if (!payloadPath && !url) {
  console.error('Usage: node scripts/linear-brief.mjs (--payload <issue.json> | --url <linear-issue-url>) [--api] [--dry-run]');
  process.exit(1);
}

function extractKeyFromUrl(inputUrl) {
  if (!inputUrl) return undefined;
  const m = inputUrl.match(/\/issue\/([A-Z]+-\d+)/i);
  return m?.[1]?.toUpperCase();
}

function normalizeIssue(raw) {
  if (!raw || typeof raw !== 'object') return {};
  const data = raw.data ?? raw;
  return {
    key: data.identifier ?? data.key ?? data.id ?? '',
    title: data.title ?? '',
    description: data.description ?? '',
    state: data.state?.name ?? data.state ?? 'Todo',
    priority: data.priorityLabel ?? data.priority ?? 'Unspecified',
    url: data.url ?? '',
  };
}

async function fetchIssueByKey(key) {
  const token = process.env.LINEAR_API_KEY;
  if (!token) {
    throw new Error('LINEAR_API_KEY is required for --api mode');
  }
  const query = `
    query($key: String!) {
      issues(filter: { identifier: { eq: $key } }, first: 1) {
        nodes {
          identifier
          title
          description
          url
          priorityLabel
          state { name }
        }
      }
    }
  `;

  const res = await fetch('https://api.linear.app/graphql', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      Authorization: token,
    },
    body: JSON.stringify({ query, variables: { key } }),
  });

  if (!res.ok) {
    throw new Error(`Linear API error: HTTP ${res.status}`);
  }

  const body = await res.json();
  if (body.errors?.length) {
    throw new Error(`Linear API GraphQL error: ${body.errors[0].message}`);
  }

  const issue = body?.data?.issues?.nodes?.[0];
  if (!issue) {
    throw new Error(`No issue found for key ${key}`);
  }

  return normalizeIssue(issue);
}

function briefTemplate(issue) {
  const key = issue.key || 'UNKNOWN-KEY';
  const title = issue.title || 'Untitled ticket';
  const state = issue.state || 'Todo';
  const priority = issue.priority || 'Unspecified';
  const issueUrl = issue.url || url || '(not provided)';
  const description = issue.description ? issue.description.trim() : '(no description)';

  return `[Ticket]\nLinear: ${key} (${issueUrl})\nTitle: ${title}\nState: ${state}\nPriority: ${priority}\n\n[Context]\n${description}\n\n[Outcome]\n- Define concrete user-visible result\n\n[Scope]\nIn:\n- \nOut:\n- \n\n[Implementation constraints]\n- Keep changes minimal and reversible\n- No fake integrations; document required env vars\n\n[Validation]\nRun and pass:\n- make lint\n- make test\n- make build\n\n[Handoff]\nReturn:\n- summary\n- changed files\n- risks/follow-ups\n- commit hash\n`;
}

(async () => {
  try {
    let issue = {};

    if (payloadPath) {
      const raw = JSON.parse(fs.readFileSync(payloadPath, 'utf8'));
      issue = normalizeIssue(raw);
      if (!issue.key && url) issue.key = extractKeyFromUrl(url);
      if (!issue.url && url) issue.url = url;
    } else {
      const key = extractKeyFromUrl(url);
      issue = { key, url };
      if (!key) {
        throw new Error('Could not extract issue key from URL. Expected .../issue/ABC-123/...');
      }
      if (useApi && !dryRun) {
        issue = await fetchIssueByKey(key);
      }
    }

    if (dryRun) {
      console.error('[dry-run] Skipping Linear API calls.');
    }

    process.stdout.write(briefTemplate(issue));
  } catch (err) {
    console.error(`linear-brief error: ${err.message}`);
    process.exit(1);
  }
})();
