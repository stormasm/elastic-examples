# elastic-examples

### NOTE

This repo has no association with Redis or pulling
data out of Redis...

If you want examples combining Redis and Elastic see
[ElasticHacker](https://github.com/stormasm/elastichacker)

```
go build -o bulkinsert bulkinsert.go
./bulkinsert -index raton -type peter -n 100 -bulk-size 10
./bulkinsert -index=warehouse -type=product -n=100 -bulk-size=10
```
##### Legacy no longer used...

Remove this shortly as it is no longer the way things work...

```
cd src/github.com/olivere
git clone git@github.com:stormasm/elastic-examples.git
cd elastic-examples/bulkstring
go get golang.org/x/sync/errgroup
go install
bulkstring -index=warehouse -type=product -n=100 -bulk-size=10
```
