# 🗃 Examples

This directory contains two minimal bootstrap examples generated from `osixia/baseimage`.

## Available examples

### `single-process`

A minimal image with one managed service.

Use it to see the simplest service lifecycle supported by the baseimage:

- one service directory
- one foreground process
- basic install, startup, and finish hooks

### `multi-process`

A minimal image with two managed services in the same image.

Use it to see how the baseimage manages multiple services concurrently:

- multiple service directories
- multiple managed processes
- per-service lifecycle hooks

## Generate examples

### single-process
```
mkdir single-process
```
```
docker run --rm --user $UID --volume $(pwd)/single-process:/run/container/generator osixia/baseimage generate bootstrap --image osixia/single-process-example --from-image osixia/baseimage
```

### multi-process
```
mkdir multi-process
```
```
docker run --rm --user $UID --volume $(pwd)/multi-process:/run/container/generator osixia/baseimage generate bootstrap --image osixia/multi-process-example --from-image osixia/baseimage --multi-process
```
