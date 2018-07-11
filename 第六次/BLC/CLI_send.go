package BLC

func (cli *QYH_CLI) QYH_send(from []string, to []string, amount []string) {
	bc := QYH_NewBlockchain()
	defer bc.qyh_db.Close()
	bc.QYH_MineNewBlock(from, to, amount)
}
