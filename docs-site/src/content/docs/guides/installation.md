---
title: Installation
description: Install the aic CLI via curl, npm, GitHub Releases, or go install.
---

`aic` ships as a single static Go binary. Pick whichever method fits your setup.

## curl | sh (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh | sh
```

Pin a specific version:

```bash
AIC_VERSION=v0.1.0 curl -sSL https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh | sh
```

## npm

```bash
npm install -g @runaic/aic
```

## GitHub Releases

Download the binary for your OS and architecture from the
[Releases page](https://github.com/learners-superpumped/aic/releases).

## go install

```bash
go install github.com/learners-superpumped/aic@latest
```

## Verify

```bash
aic --version
```

Next: [Quick start](/guides/quick-start/).
