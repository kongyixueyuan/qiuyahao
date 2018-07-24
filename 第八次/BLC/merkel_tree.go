package BLC

import "crypto/sha256"

type QYH_MerkelTree struct {
	QYH_RootNode *QYH_MerkelNode
}

type QYH_MerkelNode struct {
	QYH_Left  *QYH_MerkelNode
	QYH_Right *QYH_MerkelNode
	QYH_Data  []byte
}

func QYH_NewMerkelTree(data [][]byte) *QYH_MerkelTree {
	var nodes []QYH_MerkelNode

	// 如果交易数据不是双数，将最后一个交易复制添加到最后
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}
	// 生成所有的一级节点，存储到node中
	for _, dataum := range data {
		node := QYH_NewMerkelNode(nil, nil, dataum)
		nodes = append(nodes, *node)
	}

	// 遍历生成顶层节点
	for i := 0;i<len(data)/2 ;i++{
		var newLevel []QYH_MerkelNode
		for j:=0 ; j<len(nodes) ;j+=2  {
			node := QYH_NewMerkelNode(&nodes[j],&nodes[j+1],nil)
			newLevel = append(newLevel,*node)
		}
		nodes = newLevel
	}

	//for ; len(nodes)==1 ;{
	//	var newLevel []Rwq_MerkelNode
	//	for j:=0 ; j<len(nodes) ;j+=2  {
	//		node := Rwq_NewMerkelNode(&nodes[j],&nodes[j+1],nil)
	//		newLevel = append(newLevel,*node)
	//	}
	//	nodes = newLevel
	//}
	mTree := QYH_MerkelTree{&nodes[0]}
	return &mTree
}

// 新叶节点
func QYH_NewMerkelNode(left, right *QYH_MerkelNode, data []byte) *QYH_MerkelNode {
	mNode := QYH_MerkelNode{}

	if left == nil && right == nil {
		hash := sha256.Sum256(data)
		mNode.QYH_Data = hash[:]
	} else {
		prevHashes := append(left.QYH_Data, right.QYH_Data...)
		hash := sha256.Sum256(prevHashes)
		mNode.QYH_Data = hash[:]
	}

	mNode.QYH_Left = left
	mNode.QYH_Right = right

	return &mNode
}
