package BLC

import (
	"fmt"
)

func (cli *CLI) CreateWallet()  {

	wallets, _ := NewWallets()

	wallets.CreateNewWallet()

	fmt.Println(wallets.WalletsMap)
}
