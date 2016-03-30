Purpose
---
`fan` is a command line utility that splits input to n workers, line by line, and collects their output, line by line. It does a fan-out fan-in, reading from stdin and writing to stdout. 

Install from source
---
```sh
make install
```

Install binaries
---
Change `darwin` to `linux` as needed in the URL:
```sh
curl -Os https://storage.googleapis.com/peakunicorn/bin/amd64/darwin/fan
chmod 0755 fan
sudo mv fan /usr/local/bin/fan
```

Use
---
```sh
# have 4 'myscript' workers process 'myfile' in parallel
cat myfile | fan -n=4 myscript >out
# equivalent
fan -n=4 myscript <myfile >out
# equivalent
fan -n=4 <myfile myscript >out
```

```sh
# same, but also tee the individual workers' inputs ($$ is the worker PID)
cat myfile | fan -n=4 bash -c 'tee out.$$ | myscript' >out
```
