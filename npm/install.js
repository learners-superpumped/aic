#!/usr/bin/env node
const fs = require("fs");
const path = require("path");
const https = require("https");
const { execSync } = require("child_process");
const zlib = require("zlib");
const tar = require("tar");

const pkg = require("./package.json");
const REPO = "learners-superpumped/aic";
const VERSION = `v${pkg.version}`;
const BIN_DIR = path.join(__dirname, "bin");

function platformTarget() {
  const os = { darwin: "darwin", linux: "linux" }[process.platform];
  const arch = { x64: "x86_64", arm64: "arm64" }[process.arch];
  if (!os || !arch) {
    throw new Error(`unsupported platform: ${process.platform}/${process.arch}`);
  }
  return { os, arch };
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { "User-Agent": "aic-npm-installer" } }, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return download(res.headers.location, dest).then(resolve, reject);
      }
      if (res.statusCode !== 200) {
        return reject(new Error(`HTTP ${res.statusCode} from ${url}`));
      }
      const file = fs.createWriteStream(dest);
      res.pipe(file);
      file.on("finish", () => file.close(resolve));
      file.on("error", reject);
    }).on("error", reject);
  });
}

async function main() {
  const { os, arch } = platformTarget();
  const archive = `aic_${pkg.version}_${os}_${arch}.tar.gz`;
  const url = `https://github.com/${REPO}/releases/download/${VERSION}/${archive}`;
  const tmpArchive = path.join(BIN_DIR, archive);

  fs.mkdirSync(BIN_DIR, { recursive: true });
  console.log(`aic: downloading ${url}`);
  await download(url, tmpArchive);

  await tar.x({ file: tmpArchive, cwd: BIN_DIR, filter: (p) => p === "aic" });
  fs.chmodSync(path.join(BIN_DIR, "aic"), 0o755);
  fs.unlinkSync(tmpArchive);
  console.log("aic: install complete");
}

main().catch((err) => {
  console.error("aic: install failed:", err.message);
  process.exit(1);
});
