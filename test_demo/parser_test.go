package bench

import (
	"github.com/VertexC/logparser-go/parser"
	"testing"
)

func BenchmarkLogSig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var parser parser.LogSig
		inputDir := "../logs/HDFS/"
		outputDir := "LogSig_result/"
		logFile := "HDFS_2k.log"
		logFormat := "<Date> <Time> <Pid> <Level> <Component>: <Content>"
		regexList := []string{}
		groupNum := 14
		parser.Init(inputDir, outputDir, logFile, logFormat, regexList, groupNum)
		parser.Parse()
	}
}
