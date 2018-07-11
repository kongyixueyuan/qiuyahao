package BLC

func (cli *QYH_CLI) QYH_printutxo() {
	bc := QYH_NewBlockchain()
	UTXOSet := QYH_UTXOSet{bc}
	defer bc.qyh_db.Close()
	UTXOSet.String()
}
