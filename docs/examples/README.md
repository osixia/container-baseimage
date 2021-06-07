# 🗃 Examples

## Generate examples

### single-process
```
mkdir single-process
```
```
docker run --rm --user $UID --volume $(pwd)/single-process:/run/container/generator osixia/baseimage generate bootstrap
```

### multiprocess
```
mkdir multiprocess
```
```
docker run --rm --user $UID --volume $(pwd)/multiprocess:/run/container/generator osixia/baseimage generate bootstrap --multiprocess
```
