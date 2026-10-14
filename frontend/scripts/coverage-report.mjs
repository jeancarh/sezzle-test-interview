import { readFile, writeFile } from "node:fs/promises";

const root = new URL("../../", import.meta.url);
const backendText = await readFile(new URL("coverage/backend-packages.txt", root), "utf8");
const frontendSummary = JSON.parse(await readFile(new URL("coverage/frontend/coverage-summary.json", root), "utf8"));

const modulePrefix = "github.com/jeancarh/sezzle-test-interview/backend/";
const backendRows = backendText
  .split("\n")
  .map((line) => /^ok\s+(\S+)\s+.*coverage: ([\d.]+)% of statements/.exec(line.trim()))
  .filter(Boolean)
  .map(([, pkg, pct]) => `| \`${pkg.replace(modulePrefix, "")}\` | ${pct}% |`);

const percent = (value) => `${value.pct.toFixed(1)}%`;
const frontendRoot = new URL("frontend/", root).pathname;
const frontendRows = Object.entries(frontendSummary)
  .filter(([file]) => file !== "total")
  .map(([file, m]) => `| \`${file.replace(frontendRoot, "")}\` | ${percent(m.statements)} | ${percent(m.branches)} | ${percent(m.functions)} | ${percent(m.lines)} |`)
  .sort();
const total = frontendSummary.total;

const report = `# Coverage

Generated with \`make coverage\`. HTML reports are written to \`coverage/backend.html\` and \`coverage/frontend/index.html\` (not committed). CI uploads the same reports as build artifacts.

## Backend (Go, statement coverage)

| Package | Coverage |
| --- | --- |
${backendRows.join("\n")}

\`cmd/api\` and the embed-only packages have no tests: they only wire things together, and the integration tests exercise the same router that \`main.go\` builds.

## Frontend (Vitest, v8)

| File | Statements | Branches | Functions | Lines |
| --- | --- | --- | --- | --- |
${frontendRows.join("\n")}
| **Total** | **${percent(total.statements)}** | **${percent(total.branches)}** | **${percent(total.functions)}** | **${percent(total.lines)}** |
`;

await writeFile(new URL("COVERAGE.md", root), report);
