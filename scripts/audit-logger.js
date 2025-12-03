#!/usr/bin/env node

/**
 * Audit Logger - Tracks all code changes in the repository
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const AUDIT_LOG_DIR = path.join(__dirname, '..', 'audit-logs');
const AUDIT_LOG_FILE = path.join(AUDIT_LOG_DIR, 'changes.log');
const AUDIT_JSON_FILE = path.join(AUDIT_LOG_DIR, 'changes.json');

// Ensure audit log directory exists
if (!fs.existsSync(AUDIT_LOG_DIR)) {
    fs.mkdirSync(AUDIT_LOG_DIR, { recursive: true });
}

// Initialize JSON log if it doesn't exist
if (!fs.existsSync(AUDIT_JSON_FILE)) {
    fs.writeFileSync(AUDIT_JSON_FILE, JSON.stringify({ changes: [] }, null, 2));
}

function getGitInfo() {
    try {
        const branch = execSync('git rev-parse --abbrev-ref HEAD', { encoding: 'utf-8' }).trim();
        const commitHash = execSync('git rev-parse HEAD', { encoding: 'utf-8' }).trim();
        const author = execSync('git config user.name', { encoding: 'utf-8' }).trim();
        const email = execSync('git config user.email', { encoding: 'utf-8' }).trim();
        return { branch, commitHash, author, email };
    } catch (error) {
        return { branch: 'unknown', commitHash: 'unknown', author: 'unknown', email: 'unknown' };
    }
}

function getChangedFiles() {
    try {
        // Get staged files
        const staged = execSync('git diff --cached --name-status', { encoding: 'utf-8' });
        const files = [];

        staged.split('\n').forEach(line => {
            if (line.trim()) {
                const [status, ...pathParts] = line.split('\t');
                const filePath = pathParts.join('\t');
                let changeType = 'modified';

                if (status.startsWith('A')) changeType = 'added';
                else if (status.startsWith('D')) changeType = 'deleted';
                else if (status.startsWith('M')) changeType = 'modified';
                else if (status.startsWith('R')) changeType = 'renamed';

                files.push({ path: filePath, type: changeType, status });
            }
        });

        return files;
    } catch (error) {
        return [];
    }
}

function getDiff(filePath) {
    try {
        const diff = execSync(`git diff --cached -- "${filePath}"`, { encoding: 'utf-8' });
        return diff || 'Binary file or no diff available';
    } catch (error) {
        return 'Error retrieving diff';
    }
}

function logChange() {
    const timestamp = new Date().toISOString();
    const gitInfo = getGitInfo();
    const changedFiles = getChangedFiles();

    if (changedFiles.length === 0) {
        console.log('No changes to log');
        return;
    }

    // Prepare audit entry
    const auditEntry = {
        timestamp,
        gitInfo,
        files: changedFiles.map(file => ({
            ...file,
            diff: getDiff(file.path)
        }))
    };

    // Write to JSON log
    const jsonData = JSON.parse(fs.readFileSync(AUDIT_JSON_FILE, 'utf-8'));
    jsonData.changes.push(auditEntry);
    fs.writeFileSync(AUDIT_JSON_FILE, JSON.stringify(jsonData, null, 2));

    // Write to text log
    const logEntry = `
${'='.repeat(80)}
TIMESTAMP: ${timestamp}
BRANCH: ${gitInfo.branch}
COMMIT: ${gitInfo.commitHash}
AUTHOR: ${gitInfo.author} <${gitInfo.email}>
${'='.repeat(80)}

CHANGED FILES:
${changedFiles.map(f => `  ${f.type.toUpperCase()}: ${f.path}`).join('\n')}

${changedFiles.map(f => `
--- FILE: ${f.path} (${f.type}) ---
${f.diff}
`).join('\n')}

`;

    fs.appendFileSync(AUDIT_LOG_FILE, logEntry);

    console.log(`✓ Audit log updated: ${changedFiles.length} file(s) logged`);
}

// Run the logger
logChange();