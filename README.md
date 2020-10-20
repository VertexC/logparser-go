[![Build Status](https://travis-ci.org/VertexC/logparser-go.svg?branch=master)](https://travis-ci.org/VertexC/logparser-go)

# LogParser In Golang
This project is originated from [logpari/logparser](https://github.com/logpai/logparser), which privides toolkit and benchmars for automated log parsing implemented in python. 

The motivation of this project is to re-implemented algorithm in golang for run-time efficiency.

## Log Dataset
All datasets under `logs` are from [logpari/logparser](https://github.com/logpai/logparser).

## Usage
### Benchmark
```bash
~/b/p/logparser-go (master)> go run main.go
Start to Run LogSig on:  HDFS
{0.8874827 0.9889668 0.9354805 0.375}
Time duration:  1.21992258s
Start to Run LogSig on:  Hadoop
{0.9996366 0.90963674 0.9525154 0.4665}
Time duration:  3.183971931s
```
### Unit Test`
The folder `test` consists of demos which run HDFS data set and produce corresponding templates and structed result under `test` folder.

```bash
go test -v test/*.go
```
