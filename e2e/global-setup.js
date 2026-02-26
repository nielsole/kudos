const { execSync, spawn } = require('child_process');
const path = require('path');
const http = require('http');
const fs = require('fs');

const ROOT = path.resolve(__dirname, '..');
const BACKEND_PID_FILE = path.join(__dirname, '.kudos-e2e-backend.pid');
const REDIS_PID_FILE = path.join(__dirname, '.kudos-e2e-redis.pid');
const BINARY = path.join(ROOT, 'kudos-e2e-binary');
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
  // ── Redis ────────────────────────────────────────────────────────────────
  // Prefer a local redis-server binary; fall back to Docker.
  const hasRedisServer = (() => {
    try { execSync('which redis-server', { stdio: 'ignore' }); return true; } catch (_) { return false; }
  })();

  if (hasRedisServer) {
    console.log('[e2e setup] Starting Redis (redis-server)...');
    // Shut down any stale instance on our test port
    try { execSync(`redis-cli -p ${REDIS_PORT} shutdown nosave`, { stdio: 'ignore' }); } catch (_) {}
    await new Promise((r) => setTimeout(r, 200));

    const redisProc = spawn('redis-server', ['--port', String(REDIS_PORT), '--loglevel', 'warning'], {
      cwd: ROOT,
      stdio: ['ignore', 'pipe', 'pipe'],
      detached: false,
    });
    redisProc.stderr.on('data', (d) => process.stderr.write(`[redis] ${d}`));
    fs.writeFileSync(REDIS_PID_FILE, String(redisProc.pid));
    // Wait for redis to accept connections
    await waitForRedis(REDIS_PORT);
  } else {
    console.log('[e2e setup] Starting Redis (Docker)...');
    const CONTAINER = 'kudos-e2e-redis';
    execSync(`docker rm -f ${CONTAINER} 2>/dev/null || true`, { stdio: 'ignore' });
    execSync(`docker run --name ${CONTAINER} -d -p ${REDIS_PORT}:6379 redis`, { stdio: 'inherit' });
    await new Promise((r) => setTimeout(r, 1000));
  }

  // ── Go backend ───────────────────────────────────────────────────────────
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

  fs.writeFileSync(BACKEND_PID_FILE, String(server.pid));

  console.log(`[e2e setup] Waiting for backend on port ${BACKEND_PORT}...`);
  await waitForServer('localhost', BACKEND_PORT, '/?url=healthcheck');
  console.log('[e2e setup] Backend ready.');
};

// Poll redis-cli ping until Redis is ready
async function waitForRedis(port, timeoutMs = 10000) {
  const start = Date.now();
  while (Date.now() - start < timeoutMs) {
    try {
      execSync(`redis-cli -p ${port} ping`, { stdio: 'ignore' });
      return;
    } catch (_) {
      await new Promise((r) => setTimeout(r, 200));
    }
  }
  throw new Error(`Redis on port ${port} did not start within ${timeoutMs}ms`);
}
