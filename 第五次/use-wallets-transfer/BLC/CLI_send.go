package BLC

import (
	"fmt"
	"os"
)

// 转账
func (cli *CLI) Send(from []string, to []string, amount []string)  {

	if DBExists() == false {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}

	blockchain := BlockchainObject()
	defer blockchain.DB.Close()

	blockchain.MineNewBlock(from, to, amount)

}
