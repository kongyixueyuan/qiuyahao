package BLC

import "fmt"

// 先用它去查询余额
func (cli *QYH_CLI) QYH_getBalance(address string,nodeID string)  {

	fmt.Println("地址：" + address)

	// 获取某一个节点的blockchain对象
	blockchain := QYH_BlockchainObject(nodeID)
	defer blockchain.QYH_DB.Close()

	utxoSet := &QYH_UTXOSet{blockchain}

	amount := utxoSet.QYH_GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n",address,amount)

}
