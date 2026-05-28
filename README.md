# aic — Run AI Company CLI

`aic` is the official command-line interface for [Run AI Company](https://runaic.com).
It is built for both humans and AI agents: a single static Go binary, OIDC browser
login for humans, and (coming soon) team-scoped API keys for machines.

## Install

### `curl | sh` (recommended)

```bash
curl -sSL https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh | sh
```

Pin a specific version:

```bash
AIC_VERSION=v0.1.0 curl -sSL https://raw.githubusercontent.com/learners-superpumped/aic/main/install.sh | sh
```

### npm

```bash
npm install -g @runaic/aic
```

### GitHub Releases

Download the binary for your OS/arch from the [Releases page](https://github.com/learners-superpumped/aic/releases).

### Go

```bash
go install github.com/learners-superpumped/aic@latest
```

## Quick start

```bash
aic login                    # OIDC browser login
aic teams list
aic projects list --team <team-id>
aic domains search example.com --team <team-id>
```

See `aic --help` and `aic <command> --help` for the full command reference.

## Source

This repository is a read-only mirror published from the private
`aicompany-platform` monorepo via splitsh-lite. Issues and pull requests are
welcome here; merged changes are upstreamed to the monorepo.

## License

MIT — see [LICENSE](./LICENSE).