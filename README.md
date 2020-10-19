# LogParser In Golang
This project is originated from [logpari/logparser](https://github.com/logpai/logparser), which privides toolkit and benchmars for automated log parsing implemented in python. 

The motivation of this project is to re-implemented algorithm in golang for run-time efficiency.

## Log Dataset
All datasets under `logs` are from [logpari/logparser](https://github.com/logpai/logparser).

## Benchmark
### Demo
The folder `test_demo` consists of demos which run HDFS data set and produce corresponding templates and structed result.

```bash
go test test_demo/*.go -bench=.
```