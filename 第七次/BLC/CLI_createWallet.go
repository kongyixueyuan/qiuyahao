package BLC

import "fmt"

func (cli *QYH_CLI) QYH_createWallet(nodeID string)  {

	wallets,_ := QYH_NewWallets(nodeID)

	wallets.QYH_CreateNewWallet(nodeID)

	fmt.Println(len(wallets.QYH_WalletsMap))
}
