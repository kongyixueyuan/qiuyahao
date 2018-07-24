package BLC

import "fmt"

func (cli *QYH_CLI) QYH_createWallet(nodeID string) {
	//wallet := Rwq_NewWallet()
	//address := wallet.Rwq_GetAddress()
	//fmt.Printf("钱包地址：%s\n",address)

	wallets, _ := QYH_NewWallets(nodeID)
	address := wallets.QYH_NewWallet().QYH_GetAddress()
	wallets.QYH_SaveToFile(nodeID)
	fmt.Printf("钱包地址：%s\n", address)

}
