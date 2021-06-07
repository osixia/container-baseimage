# 🗃 Examples

## Generate examples

### single-process
```
mkdir single-process
```
```
docker run --rm --user $UID --volume $(pwd)/single-process:/container/generator/output osixia/baseimage generate bootstrap
```

### multiprocess
```
mkdir multiprocess
```
```
docker run --rm --user $UID --volume $(pwd)/multiprocess:/container/generator/output osixia/baseimage generate bootstrap --multiprocess
```
