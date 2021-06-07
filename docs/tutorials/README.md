# 🎮 Tutorials

This directory contains small, implementation-driven tutorials built on top of `osixia/baseimage`.

## Available Tutorials

### `nginx`

Build a single-service Nginx image with:

- package installation in the Dockerfile
- a service install step for static site initialization
- a startup step driven by environment variables
- an Nginx foreground process managed by the baseimage entrypoint

### `nginx-php-fpm`

Extend the Nginx tutorial with a second managed service:

- Nginx continues to serve HTTP traffic
- PHP-FPM runs as a separate managed process
- communication happens through a Unix socket in `/run/container/php-fpm.sock`

## Notes

- These tutorials are examples of how the current codebase is intended to be used.
- They are small on purpose and focus on service lifecycle structure more than production hardening.
- Read the main [documentation](../README.md) for runtime details such as lifecycle flags, logging, restart behavior, and read-only/rootless constraints.
