# Changelog

## Unreleased

## 2.1.0

### Added

- Create `/run/container/tmp` as a shared temporary directory under the configured
  runtime root, without changing `TMPDIR`.

### Fixed

- Build with Go 1.26.10 to include the crypto/x509 certificate verification
  security fix (GO-2026-6809), and align both modules and CI with that toolchain.

- Detect watched files retargeted by sibling symlink replacements, including
  Kubernetes Secret and ConfigMap `..data` updates.

## 2.0.0

Originally released as osixia/light-baseimage on July 23, 2015.

🚀 osixia/baseimage version 2.0.0 is a complete rewrite in Go, built on more than 10 years of real-world experience from version 1.
