package BLC

import "fmt"

// 转账
func (cli *QYH_CLI) QYH_send(from []string,to []string,amount []string,nodeID string, mineNow bool)  {


	blockchain := QYH_BlockchainObject(nodeID)
	defer blockchain.QYH_DB.Close()

	if mineNow {
		blockchain.QYH_MineNewBlock(from,to,amount,nodeID)

		utxoSet := &QYH_UTXOSet{blockchain}

		//转账成功以后，需要更新一下
		utxoSet.QYH_Update()
	} else {
		// 把交易发送到矿工节点去进行验证
		fmt.Println("由矿工节点处理......")
	}



}

