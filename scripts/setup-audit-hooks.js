#!/usr/bin/env node

/**
 * Setup script for audit logging Git hooks
 */

const fs = require('fs');
const path = require('path');

const GIT_HOOKS_DIR = path.join(__dirname, '..', '.git', 'hooks');
const PRE_COMMIT_HOOK = path.join(GIT_HOOKS_DIR, 'pre-commit');
const POST_COMMIT_HOOK = path.join(GIT_HOOKS_DIR, 'post-commit');

// Ensure hooks directory exists
if (!fs.existsSync(GIT_HOOKS_DIR)) {
  console.error('Error: .git/hooks directory not found. Are you in a Git repository?');
  process.exit(1);
}

// Pre-commit hook content
const preCommitContent = `#!/bin/bash
# Audit Logger - Pre-commit Hook

node "$(dirname "$0")/../../scripts/audit-logger.js"
`;

// Post-commit hook content
const postCommitContent = `#!/bin/bash
# Audit Logger - Post-commit Hook

echo "✓ Change logged to audit-logs/changes.log"
`;

function createHook(hookPath, content) {
  if (fs.existsSync(hookPath)) {
    console.log(`⚠ Hook already exists: ${path.basename(hookPath)}`);
    console.log('  Backing up existing hook...');
    fs.copyFileSync(hookPath, `${hookPath}.backup`);
  }
  
  fs.writeFileSync(hookPath, content, { mode: 0o755 });
  console.log(`✓ Created hook: ${path.basename(hookPath)}`);
}

// Create hooks
console.log('\n🔧 Setting up audit logging hooks...\n');
createHook(PRE_COMMIT_HOOK, preCommitContent);
createHook(POST_COMMIT_HOOK, postCommitContent);

// Create audit logs directory
const auditLogsDir = path.join(__dirname, '..', 'audit-logs');
if (!fs.existsSync(auditLogsDir)) {
  fs.mkdirSync(auditLogsDir, { recursive: true });
  console.log('✓ Created audit-logs directory');
}

// Create initial log files
const changesLog = path.join(auditLogsDir, 'changes.log');
const changesJson = path.join(auditLogsDir, 'changes.json');

if (!fs.existsSync(changesLog)) {
  fs.writeFileSync(changesLog, `Audit Log - Initialized on ${new Date().toISOString()}\n${'='.repeat(80)}\n\n`);
  console.log('✓ Initialized changes.log');
}

if (!fs.existsSync(changesJson)) {
  fs.writeFileSync(changesJson, JSON.stringify({ changes: [] }, null, 2));
  console.log('✓ Initialized changes.json');
}

// Create README
const readmePath = path.join(auditLogsDir, 'README.md');
const readmeContent = `# Audit Logs

This directory contains automatic logs of all code changes made to the repository.

## Files

- **changes.log**: Human-readable text log with detailed diffs
- **changes.json**: Machine-readable JSON log for programmatic access

## How It Works

The audit logger is triggered automatically by Git hooks:
- **pre-commit**: Captures changes before they are committed
- **post-commit**: Confirms the logging operation

## Viewing Logs

### Text Format
\`\`\`bash
cat audit-logs/changes.log
# or
less audit-logs/changes.log
\`\`\`

### JSON Format
\`\`\`bash
cat audit-logs/changes.json | jq '.'
\`\`\`

## Filtering Logs

### By Date
\`\`\`bash
grep "TIMESTAMP:" audit-logs/changes.log | grep "2025-12"
\`\`\`

### By Author
\`\`\`bash
grep "AUTHOR:" audit-logs/changes.log | grep "YourName"
\`\`\`

### By File
\`\`\`bash
grep "FILE:" audit-logs/changes.log | grep "package.json"
\`\`\`

## Maintenance

These logs are **not committed to Git** by default (see .gitignore).
Archive old logs periodically to keep the directory size manageable.

---
Generated: ${new Date().toISOString()}
`;

fs.writeFileSync(readmePath, readmeContent);
console.log('✓ Created README.md');

console.log('\n✅ Audit logging setup complete!\n');
console.log('📝 Changes will now be automatically logged to audit-logs/\n');
