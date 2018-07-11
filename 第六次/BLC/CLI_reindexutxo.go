package BLC

import "fmt"

func (cli *QYH_CLI) QYH_reindexUTXO() {
	bc := QYH_NewBlockchain();
	defer bc.qyh_db.Close()
	utxoset := QYH_UTXOSet{bc}
	utxoset.QYH_Reindex()
	fmt.Println("重建成功")
}
