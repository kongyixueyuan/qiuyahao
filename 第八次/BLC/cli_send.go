package BLC

func (cli *QYH_CLI) QYH_send(from []string, to []string, amount []string,nodeID string, mineNow bool) {
	bc := QYH_NewBlockchain(nodeID)
	defer bc.QYH_db.Close()
	bc.QYH_MineNewBlock(from, to, amount,nodeID, mineNow)
}