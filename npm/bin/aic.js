#!/usr/bin/env node
const { spawnSync } = require("child_process");
const path = require("path");

const binary = path.join(__dirname, "aic");
const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
process.exit(result.status ?? 1);
