package parser

import (
	"fmt"
	"github.com/gonum/stat/combin"
)

type Metric struct {
	precision float32
	recall    float32
	fMeasure  float32
	accuracy  float32
}

func Evaluate(gtFilePath string, resultFilePath string) {
	gtDataFrame := ReadCsvFile(gtFilePath) // groud truth data frame
	resultDataFrame := ReadCsvFile(resultFilePath)
	eventGt := gtDataFrame.records["EventId"]
	eventResult := resultDataFrame.records["EventId"]
	metric := getMetric(eventGt, eventResult)
	fmt.Println(metric)
}

func GroupArrayIndexByValue(array []string) map[string][]int {
	result := map[string][]int{}
	for idx, val := range array {
		if _, ok := result[val]; ok {
			result[val] = append(result[val], idx)
		} else {
			result[val] = []int{idx}
		}
	}
	return result
}

func combCounts(cluster map[string][]int) int {
	counts := 0
	for _, logIds := range cluster {
		if len(logIds) > 1 {
			counts += combin.Binomial(len(logIds), 2)
		}
	}
	return counts
}

func getMetric(eventGt []string, eventResult []string) Metric {
	// group by eventId
	gtClusters := GroupArrayIndexByValue(eventGt)
	resClusters := GroupArrayIndexByValue(eventResult)

	// get comb counts
	resCombCounts := combCounts(resClusters)
	gtCombCoutns := combCounts(gtClusters)

	accurateEvents := 0
	hitCombCounts := 0
	for _, logIds := range resClusters {
		gt := map[string][]int{} // eventId: logIds
		for _, logId := range logIds {
			gtCId := eventGt[logId]
			if _, ok := gt[gtCId]; ok {
				gt[gtCId] = append(gt[gtCId], logId)
			} else {
				gt[gtCId] = []int{logId}
			}
		}
		if len(gt) == 1 {
			for gtCId := range gt {
				if len(gtClusters[gtCId]) == len(logIds) {
					accurateEvents += len(logIds)
				}
				break
			}
		}
		hitCombCounts += combCounts(gt)
	}

	precision := float32(hitCombCounts) / float32(resCombCounts)
	recall := float32(hitCombCounts) / float32(gtCombCoutns)

	fMeasure := 2 * precision * recall / (precision + recall)
	accuracy := float32(accurateEvents) / float32(len(eventGt))
	return Metric{precision, recall, fMeasure, accuracy}
}
