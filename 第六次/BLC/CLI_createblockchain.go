package BLC

import "log"

func (cli *QYH_CLI) QYH_createblockchain(address string)  {
	//验证地址是否有效
	if !QYH_ValidateAddress(address){
		log.Panic("地址无效")
	}
	bc := QYH_CreateBlockchain(address)
	defer bc.qyh_db.Close()

	// 生成UTXOSet数据库
	UTXOSet := QYH_UTXOSet{bc}
	UTXOSet.QYH_Reindex()
}
