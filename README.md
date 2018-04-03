# storage-s3-migrator

Fast and correct migrator from hash storage (3level - bytes dir tree) to S3 API storage

## input directory

3 level bytes tree extracted from first 3 bytes
sha256 named files (`.dat` extension are ignored)

```
/mnt/test/
├── 00
│   ├── 00
│   │   ├── 00
│   │   │   ├── 00000002B02F565AF273B13DA8770D5E2803C1A8D8239DA29472D59857462CA9.dat
│   │   │   ├── 00000048B1C9E60C14A6619F0292DEA96DF7F10C11CFA9AE28693219C0AE844B.dat
│   │   │   ├── 0000004DDD7930A75AEEE9FE98F1EECB2BC59EB840DE20AC574507644D2A0329.dat
│   │   │   ├── 000000552925D3F0948474DD0FB116FC1BEF526A45FA6E171503C52B8CFCB732.dat
│   │   │   ├── 0000007242D7A5F1A36BFC565FA13990AFCCAD75FC28C7B6F05EAD95173478F0.dat
│   │   │   ├── 0000009E35110AE7CB267846EC7290AB2EE4ACCDF19C278071C6C4BAEB2C52B9.dat
```

## output

* STDOUT output is progressbar
* STDERR output is log (I recommend redirect log output to file e.g. `2> migration.log`)

## concurrency

first byte directory are sharded by concurrent option

each shard are concurrently uploaded

directory are uploaded randomly for maximal utilization 

## help

```
usage: storage-s3-migrator --endpoint=ENDPOINT --namespace=NAMESPACE --user=USER --pass=PASS [<flags>] <dir>

Flags:
  --help                 Show context-sensitive help (also try --help-long and --help-man).
  --concurrent=8         count of concurent uploader
  --endpoint=ENDPOINT    endpoint hostname
  --namespace=NAMESPACE  endpoint namespace
  --user=USER            username
  --pass=PASS            password
  --custom-last-modifed  set x-amz-meta-Last-Modified header with last modification time of source file

Args:
  <dir>  source directory
```
