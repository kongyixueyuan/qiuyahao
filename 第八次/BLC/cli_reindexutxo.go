package BLC

import "fmt"

func (cli *QYH_CLI) QYH_reindexUTXO(nodeID string)  {
	bc := QYH_NewBlockchain(nodeID);
	defer bc.QYH_db.Close()
	utxoset := QYH_UTXOSet{bc}
	utxoset.QYH_Reset()
	fmt.Println("重建成功")
}
