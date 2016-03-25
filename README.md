Purpose
---
`fan` is a command line utility that splits input to n workers, line by line, and collects their output, line by line. It does a fan-out fan-in, reading from stdin and writing to stdout. 

Use
---
```
cat myfile | fan -n=4 myscript >out
cat myfile | fan -n=4 bash -c 'tee out.$$' >out
```
