package BLC

import (
	"log"
	"fmt"
)

func (cli *QYH_CLI) QYH_listAddrsss() {
	wallets, err := QYH_NewWallets()

	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.QYH_GetAddresses()

	for _, address := range addresses {
		fmt.Printf("%s\n", address)
	}
}
