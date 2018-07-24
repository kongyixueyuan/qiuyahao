package BLC

func (cli *QYH_CLI) QYH_printutxo(nodeID string) {
	bc := QYH_NewBlockchain(nodeID)
	UTXOSet := QYH_UTXOSet{bc}
	defer bc.QYH_db.Close()
	UTXOSet.QYH_String()
}
