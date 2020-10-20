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
metric: {precision:0.8888562 recall:0.9975381 fMeasure:0.9400664 accuracy:0.5825}
Time duration:  1.425726157s

Start to Run LogSig on:  Hadoop
metric: {precision:0.9995276 recal
```
### Test
The folder `test` consists of demos which run HDFS data set and produce corresponding templates and structed result under `test` folder.

```bash
go test -v test/*.go
```
