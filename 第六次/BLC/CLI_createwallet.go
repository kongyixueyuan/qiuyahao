package BLC

import "fmt"

func (cli *QYH_CLI) QYH_createWallet() {

	wallets, _ := QYH_NewWallets()
	address := wallets.QYH_NewWallet().QYH_GetAddress()
	wallets.QYH_SaveToFile()
	fmt.Printf("钱包地址：%s\n", address)

}
