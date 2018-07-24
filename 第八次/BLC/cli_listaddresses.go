package BLC

import (
	"log"
	"fmt"
)

func (cli *QYH_CLI) QYH_listAddrsss(nodeID string)  {
	wallets,err := QYH_NewWallets(nodeID)

	if err!=nil{
		log.Panic(err)
	}
	addresses := wallets.QYH_GetAddresses()

	for _,address := range addresses{
		fmt.Printf("%s\n",address)
	}
}
