# REACHABLE GitHub And GitLab Demo Repositories

This file explains the public CI/CD demo repository layout. The GitHub and
GitLab sets are intentionally symmetrical: each ecosystem has a reusable
toolkit, a distribution/discovery surface, and a Go remediation/public-clone
proof demo that also supports scan-only runs with remediation disabled.

## GitHub Repositories

| Repository | Primary role | Use this when |
|---|---|---|
| [`reach-testbed-github-marketplace`](https://github.com/sthenos-security/reach-testbed-github-marketplace) | GitHub Marketplace distribution surface plus the configurable root action. | You need the public GitHub Marketplace listing or a step-level `uses: sthenos-security/reach-testbed-github-marketplace@v1` action. |
| [`reach-ci-github`](https://github.com/sthenos-security/reach-ci-github) | Reusable GitHub Actions toolkit for production auto-remediation. | You want the recommended customer workflow with branch creation, proof scan, optional PR, artifacts, and Pages proof. |
| [`reach-testbed-github-go`](https://github.com/sthenos-security/reach-testbed-github-go) | Go public-clone/remediation proof demo plus scan-only mode. | You want to prove Go findings, public source cloning, MCP GitHub cloning, git clone fallback, post-remediation proof, or a public scan-only sample with `remediate=false`. |

## GitLab Repositories

| Repository | Primary role | GitHub equivalent |
|---|---|---|
| [`reach-testbed-gitlab-catalog`](https://gitlab.com/sthenos-security-public/reach-testbed-gitlab-catalog) | GitLab CI/CD Catalog component plus full remediation demo. GitLab Catalog is the GitLab distribution surface; commercial partner routing is separate. | `reach-testbed-github-marketplace` |
| [`reach-ci-gitlab`](https://gitlab.com/sthenos-security-public/reach-ci-gitlab) | Reusable GitLab remediation toolkit. | `reach-ci-github` |
| [`reach-testbed-gitlab-go`](https://gitlab.com/sthenos-security-public/reach-testbed-gitlab-go) | Go public-clone/remediation proof demo plus scan-only mode. | `reach-testbed-github-go` |

## Architecture

The Marketplace/Catalog repositories are the discovery and onboarding
surfaces. The toolkit repositories contain the reusable CI implementation. The
testbed repositories are examples and validation targets.

Use the distribution surface first:

- GitHub: Marketplace action for discovery and step-level scan/remediation, or
  `reach-ci-github` reusable workflow for full production remediation.
- GitLab: Catalog component for the full pipeline.

Use the Go demo repos as the public sample surface for both remediation and
scan-only runs. Legacy standalone scan-only repos can be retired or kept
private without changing the supported public contract.
