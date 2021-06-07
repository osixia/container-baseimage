# 🗃 Examples

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
