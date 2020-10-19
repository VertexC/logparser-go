package parser

import (
	"fmt"
	"strconv"
	"strings"
	// "time"
)

type Node struct {
	token        string
	length       int
	children     []*Node
	clusterGroup []*Cluster
}

type Cluster struct {
	template []string
	logIds   []int
}

type Drain struct {
	parser        Parser
	simTr         float32 // similarity threshold for sequence considered to blong a cluster
	internalDepth int     // depth exclude root and leafnodes
	root          Node
	clusters      []*Cluster
}

func (model *Drain) Init(inputDir string, outputDir string, logFile string,
	logFormat string, regexList []string, simTr float32, depth int) {
	model.parser.Init(inputDir, outputDir, logFile, logFormat, regexList)
	model.simTr = simTr
	model.internalDepth = depth - 2
}

func MergeSeqs(seq1, seq2 []string) []string {
	template := []string{}
	for i, word1 := range seq1 {
		word2 := seq2[i]
		if word1 == word2 {
			template = append(template, word1)
		} else {
			template = append(template, "<*>")
		}
	}
	return template
}

func (node *Node) findLength(length int) (*Node, bool) {
	for _, child := range node.children {
		if length == child.length {
			return child, true
		}
	}
	return nil, false
}

func (node *Node) findToken(token string) (*Node, bool) {
	for _, child := range node.children {
		if token == child.token {
			return child, true
		}
	}
	return nil, false
}

func SeqSimilarity(seq1, seq2 []string) (float32, int) {
	paramNum := 0 // number of parameter <*>
	sameTokenNum := 0
	for i, token1 := range seq1 {
		token2 := seq2[i]
		if token1 == token2 {
			sameTokenNum += 1
		}
		if token1 == "<*>" {
			paramNum += 1
			continue
		}
	}
	score := float32(sameTokenNum) / float32(len(seq1))
	return score, paramNum
}

func (model *Drain) FastMatch(clusterGroup []*Cluster, wordSeq []string) *Cluster {
	var matchCluster *Cluster
	maxScore := float32(-1)
	minParamNum := -1
	for _, cluster := range clusterGroup {
		score, paramNum := SeqSimilarity(wordSeq, cluster.template)
		if score > model.simTr {
			if score > maxScore || (score == maxScore && paramNum < minParamNum) {
				maxScore = score
				minParamNum = paramNum
				matchCluster = cluster
			}
		}
	}
	return matchCluster
}

func (model *Drain) TreeSearch(wordSeq []string) *Cluster {
	var matchCluster *Cluster
	// search by length
	if child, ok := model.root.findLength(len(wordSeq)); ok {
		parent := child
		for i, word := range wordSeq {
			if i >= model.internalDepth || len(parent.clusterGroup) > 0 {
				matchCluster = model.FastMatch(parent.clusterGroup, wordSeq)
				break
			}
			if child, ok := parent.findToken(word); ok {
				parent = child
				continue
			} else {
				break
			}
		}
	}
	return matchCluster
}

func (root *Node) PrintTree(depth int) {
	info := ""
	for i := 0; i < depth; i++ {
		info += "\t"
	}
	if depth == 0 {
		info += "Root"
	} else if depth == 1 {
		info = info + "<" + strconv.Itoa(root.length) + ">"
	} else {
		info = info + "<" + root.token + ">"
	}
	fmt.Println(info)
	for _, child := range root.children {
		child.PrintTree(depth + 1)
	}
}

// TODO: merge tree construction for new cluster into tree search
func (model *Drain) AddNewCluster(wordSeq []string, cluster *Cluster) {
	model.clusters = append(model.clusters, cluster)
	seqLength := len(wordSeq)
	var lengthLayer *Node
	if child, ok := model.root.findLength(seqLength); ok {
		lengthLayer = child
	} else {
		lengthLayer = &Node{}
		lengthLayer.length = seqLength
		model.root.children = append(model.root.children, lengthLayer)
	}
	parent := lengthLayer
	// fmt.Println(lengthLayer)
	for i, word := range wordSeq {
		if i >= model.internalDepth {
			parent.clusterGroup = append(parent.clusterGroup, cluster)
			return
		}
		if child, ok := parent.findToken(word); ok {
			parent = child
			continue
		} else { // create new token
			newChild := &Node{}
			newChild.token = word
			parent.children = append(parent.children, newChild)
			parent = newChild
		}
	}
	// when length of log is smaller than internal depth
	parent.clusterGroup = append(parent.clusterGroup, cluster)
}

func (model *Drain) ConstructTree(wordSeqs [][]string) {
	for logId, wordSeq := range wordSeqs {
		matchCluster := model.TreeSearch(wordSeq)
		if matchCluster == nil { // create new cluster if no match
			newCluster := &Cluster{wordSeq, []int{logId}}
			model.AddNewCluster(wordSeq, newCluster)
		} else { // update matched cluster
			matchCluster.logIds = append(matchCluster.logIds, logId)
			matchCluster.template = MergeSeqs(matchCluster.template, wordSeq)
		}
	}
}

func (model *Drain) Parse() {
	// startTime := time.Now()
	headers, dataFrame := model.parser.LoadLog()
	wordSeqs := model.parser.GetLogContent(dataFrame)
	model.ConstructTree(wordSeqs)
	// fmt.Println(headers)
	// model.root.PrintTree(0)
	patterns := map[int]string{}
	clusterRecord := map[int]int{}
	clusterSize := map[int]int{}
	for clusterId, cluster := range model.clusters {
		patterns[clusterId] = strings.Join(cluster.template, " ")
		for _, logId := range cluster.logIds {
			clusterSize[logId] = clusterId
		}
		clusterSize[clusterId] = len(cluster.logIds)
	}
	model.parser.WriteResultToFile(headers, dataFrame, wordSeqs, patterns, clusterRecord, clusterSize, true)
	// endTime := time.Now()
	// fmt.Println("Parsing Done. Time taken: ", endTime.Sub(startTime))
}
