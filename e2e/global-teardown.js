const { execSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const ROOT = path.resolve(__dirname, '..');
const BACKEND_PID_FILE = path.join(__dirname, '.kudos-e2e-backend.pid');
const REDIS_PID_FILE = path.join(__dirname, '.kudos-e2e-redis.pid');
const BINARY = path.join(ROOT, 'kudos-e2e-binary');
const REDIS_PORT = 6380;

function killByPidFile(pidFile, label) {
  if (!fs.existsSync(pidFile)) return;
  const pid = parseInt(fs.readFileSync(pidFile, 'utf8').trim(), 10);
  if (!isNaN(pid)) {
    try {
      process.kill(pid, 'SIGTERM');
      console.log(`[e2e teardown] Killed ${label} (pid ${pid})`);
    } catch (_) {}
  }
  fs.unlinkSync(pidFile);
}

module.exports = async function globalTeardown() {
  killByPidFile(BACKEND_PID_FILE, 'kudos backend');

  // If redis-server was started by setup, kill it; otherwise remove Docker container
  if (fs.existsSync(REDIS_PID_FILE)) {
    killByPidFile(REDIS_PID_FILE, 'redis-server');
    // Give it a moment to shut down before the port is released
    await new Promise((r) => setTimeout(r, 300));
  } else {
    try {
      execSync('docker rm -f kudos-e2e-redis', { stdio: 'ignore' });
      console.log('[e2e teardown] Removed Redis container');
    } catch (_) {}
  }

  // Remove the compiled binary
  try {
    fs.unlinkSync(BINARY);
    console.log('[e2e teardown] Removed binary');
  } catch (_) {}
};
