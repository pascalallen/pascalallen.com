import { ChildProcess, spawn } from 'node:child_process';
import { existsSync } from 'node:fs';
import { createServer, Server } from 'node:http';
import { AddressInfo } from 'node:net';
import { resolve } from 'node:path';
import { MARKUP_NAME } from './fixture-data';

const repos = [
  {
    name: 'fixture-queue',
    description: 'A fixture repository',
    html_url: 'https://github.com/pascalallen/fixture-queue',
    language: 'Go',
    stargazers_count: 7,
    fork: false,
    archived: false
  },
  {
    name: MARKUP_NAME,
    description: '<script>alert(2)</script>',
    html_url: 'https://github.com/pascalallen/markup',
    language: null,
    stargazers_count: 0,
    fork: false,
    archived: false
  }
];

const packages = {
  objects: [
    {
      package: {
        name: '@pascalallen/fixture-forms',
        version: '1.2.3',
        description: 'A fixture package',
        links: { npm: 'https://www.npmjs.com/package/@pascalallen/fixture-forms' }
      }
    }
  ]
};

const listen = (server: Server): Promise<number> =>
  new Promise(done => server.listen(0, '127.0.0.1', () => done((server.address() as AddressInfo).port)));

const freePort = async (): Promise<number> => {
  const probe = createServer();
  const port = await listen(probe);
  await new Promise(done => probe.close(done));
  return port;
};

const waitFor = async (url: string, child: ChildProcess): Promise<void> => {
  const deadline = Date.now() + 15000;
  while (Date.now() < deadline) {
    if (child.exitCode !== null) {
      throw new Error(`e2e: server exited with ${child.exitCode}`);
    }
    const up = await fetch(url).then(
      response => response.ok,
      () => false
    );
    if (up) {
      return;
    }
    await new Promise(done => setTimeout(done, 100));
  }
  throw new Error(`e2e: ${url} did not come up`);
};

const startSite = async (
  binary: string,
  root: string,
  feeds: string
): Promise<{ url: string; child: ChildProcess }> => {
  const port = await freePort();
  const url = `http://127.0.0.1:${port}`;
  const child = spawn(binary, [], {
    cwd: root,
    stdio: ['ignore', 'ignore', process.env.E2E_SERVER_LOG ? 'inherit' : 'ignore'],
    env: {
      PATH: process.env.PATH ?? '',
      PORT: String(port),
      APP_ENV: 'test',
      APP_BASE_URL: url,
      GIN_MODE: 'release',
      GITHUB_TOKEN: '',
      SLACK_DM_URL: '',
      TRUSTED_PROXIES: '',
      GITHUB_REPOS_URL: `${feeds}/github`,
      NPM_SEARCH_URL: `${feeds}/npm`
    }
  });
  await waitFor(`${url}/robots.txt`, child);
  return { url, child };
};

export default async function globalSetup(): Promise<() => Promise<void>> {
  const root = resolve(__dirname, '../../..');
  const binary = resolve(process.env.E2E_SERVER ?? resolve(__dirname, '../.e2e/pascalallen'));
  if (!existsSync(binary)) {
    throw new Error(`e2e: server binary ${binary} is missing; run bin/e2e`);
  }

  const fixture = createServer((req, res) => {
    const body =
      req.url === '/ok/github' ? JSON.stringify(repos) : req.url === '/ok/npm' ? JSON.stringify(packages) : null;
    res.writeHead(body === null ? 502 : 200, { 'Content-Type': 'application/json' });
    res.end(body ?? '{}');
  });
  const fixturePort = await listen(fixture);
  const feeds = `http://127.0.0.1:${fixturePort}`;

  const ok = await startSite(binary, root, `${feeds}/ok`);
  const broken = await startSite(binary, root, `${feeds}/down`);
  process.env.E2E_BASE_URL = ok.url;
  process.env.E2E_BROKEN_URL = broken.url;

  return async () => {
    ok.child.kill();
    broken.child.kill();
    await new Promise(done => fixture.close(done));
  };
}
