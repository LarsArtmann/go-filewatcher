// Generates src/content/docs/changelog.mdx from the repository's CHANGELOG.md
// So the website changelog can never drift from the source of truth again.
// Run automatically before `astro build` and `astro dev` (see package.json).
import { readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const scriptDir = import.meta.dirname;
const sourcePath = join(scriptDir, "..", "..", "CHANGELOG.md");
const targetPath = join(scriptDir, "..", "src", "content", "docs", "changelog.mdx");

const changelog = readFileSync(sourcePath, "utf8");

// Drop the leading `# Changelog` H1 and the intro lines.
// Starlight renders its own page title from the frontmatter below.
const firstEntry = changelog.indexOf("\n## ");
if (firstEntry === -1) {
  console.error(`sync-changelog: no version entries found in ${sourcePath}`);
  process.exit(1);
}

const body = changelog.slice(firstEntry + 1).trimEnd();

const page = `---
title: Changelog
description: All notable changes to go-filewatcher, synced from CHANGELOG.md.
---

${body}
`;

writeFileSync(targetPath, page);
console.log(`sync-changelog: wrote ${targetPath} (${body.split("\n").length} lines)`);
