const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const ROOT = path.resolve(__dirname, '..');
const PID_FILE = path.join(__dirname, '.kudos-e2e.pid');
const BINARY = path.join(ROOT, 'kudos-e2e-binary');
const REDIS_CONTAINER = 'kudos-e2e-redis';

module.exports = async function globalTeardown() {
  // Kill the kudos backend process
  if (fs.existsSync(PID_FILE)) {
    const pid = parseInt(fs.readFileSync(PID_FILE, 'utf8').trim(), 10);
    if (!isNaN(pid)) {
      try {
        process.kill(pid, 'SIGTERM');
        console.log(`[e2e teardown] Killed kudos backend (pid ${pid})`);
      } catch (e) {
        // Process may have already exited
      }
    }
    fs.unlinkSync(PID_FILE);
  }

  // Remove Redis container
  try {
    execSync(`docker rm -f ${REDIS_CONTAINER}`, { stdio: 'ignore' });
    console.log('[e2e teardown] Removed Redis container');
  } catch (_) {}

  // Remove the compiled binary
  try {
    fs.unlinkSync(BINARY);
    console.log('[e2e teardown] Removed binary');
  } catch (_) {}
};
