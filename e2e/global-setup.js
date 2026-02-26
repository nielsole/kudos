const { execSync, spawn } = require('child_process');
const path = require('path');
const http = require('http');
const fs = require('fs');

const ROOT = path.resolve(__dirname, '..');
const PID_FILE = path.join(__dirname, '.kudos-e2e.pid');
const BINARY = path.join(ROOT, 'kudos-e2e-binary');
const REDIS_CONTAINER = 'kudos-e2e-redis';
const REDIS_PORT = 6380;
const BACKEND_PORT = 8082;
const ADMIN_PORT = 8083;

async function waitForServer(host, port, path, timeoutMs = 15000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      await new Promise((resolve, reject) => {
        const req = http.get({ host, port, path }, (res) => {
          res.resume();
          resolve();
        });
        req.on('error', reject);
        req.setTimeout(1000, () => {
          req.destroy();
          reject(new Error('timeout'));
        });
      });
      return; // server is ready
    } catch (_) {
      await new Promise((resolve) => setTimeout(resolve, 300));
    }
  }
  throw new Error(`Server at ${host}:${port} did not start within ${timeoutMs}ms`);
}

module.exports = async function globalSetup() {
  console.log('[e2e setup] Starting Redis...');
  execSync(`docker rm -f ${REDIS_CONTAINER} 2>/dev/null || true`, { stdio: 'ignore' });
  execSync(
    `docker run --name ${REDIS_CONTAINER} -d -p ${REDIS_PORT}:6379 redis`,
    { stdio: 'inherit' }
  );

  // Give Redis a moment to initialise before the Go server connects
  await new Promise((resolve) => setTimeout(resolve, 1000));

  console.log('[e2e setup] Building Go backend...');
  execSync(`go build -o ${BINARY}`, { cwd: ROOT, stdio: 'inherit' });

  console.log('[e2e setup] Starting kudos backend...');
  const server = spawn(BINARY, [
    '--redis-address', `localhost:${REDIS_PORT}`,
    '--port', `:${BACKEND_PORT}`,
    '--admin-port', `:${ADMIN_PORT}`,
  ], {
    cwd: ROOT,
    stdio: ['ignore', 'pipe', 'pipe'],
    detached: false,
  });

  server.stdout.on('data', (d) => process.stdout.write(`[kudos] ${d}`));
  server.stderr.on('data', (d) => process.stderr.write(`[kudos] ${d}`));

  server.on('exit', (code) => {
    if (code !== null && code !== 0) {
      console.error(`[e2e setup] kudos backend exited with code ${code}`);
    }
  });

  // Persist PID so globalTeardown can kill it
  fs.writeFileSync(PID_FILE, String(server.pid));

  console.log(`[e2e setup] Waiting for backend on port ${BACKEND_PORT}...`);
  await waitForServer('localhost', BACKEND_PORT, '/?url=healthcheck');

  console.log('[e2e setup] Backend ready.');
};
