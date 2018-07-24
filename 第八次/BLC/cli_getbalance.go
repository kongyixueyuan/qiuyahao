package BLC

import (
	"log"
	"fmt"
)

func (cli *QYH_CLI) QYH_getBalance(address string,nodeID string) {
	if !QYH_ValidateAddress(address) {
		log.Panic("错误：地址无效")
	}

	bc := QYH_NewBlockchain(nodeID)
	defer bc.QYH_db.Close()
	UTXOSet := QYH_UTXOSet{bc}

	balance := UTXOSet.QYH_GetBalance(address)
	fmt.Printf("地址:%s的余额为：%d\n", address, balance)
}

func (cli *QYH_CLI) QYH_getBalanceAll(nodeID string) {
	wallets,err := QYH_NewWallets(nodeID)
	if err!=nil{
		log.Panic(err)
	}
	balances := wallets.QYH_GetBalanceAll(nodeID)
	for address,balance := range balances{
		fmt.Printf("地址:%s的余额为：%d\n", address, balance)
	}
}