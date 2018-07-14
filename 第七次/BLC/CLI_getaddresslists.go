package BLC

import "fmt"

// 打印所有的钱包地址
func (cli *QYH_CLI) QYH_addressLists(nodeID string)  {

	fmt.Println("打印所有的钱包地址:")

	wallets,_ := QYH_NewWallets(nodeID)

	for address,_ := range wallets.QYH_WalletsMap {

		fmt.Println(address)
	}
}