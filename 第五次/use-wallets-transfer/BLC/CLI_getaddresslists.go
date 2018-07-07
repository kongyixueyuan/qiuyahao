package BLC

import "fmt"

// 打印所有钱包地址
func (cli *CLI) AddressLists() []string  {

	fmt.Println("打印所有的钱包地址：")

	wallets, _ := NewWallets()

	for address, _ :=range wallets.WalletsMap {

		fmt.Println(address)
	}

	return nil

}