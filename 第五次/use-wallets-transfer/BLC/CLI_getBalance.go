package BLC

import "fmt"

// 查询余额
func (cli *CLI) GetBalance(address string)  {

	blockchain := BlockchainObject()
	defer blockchain.DB.Close()

	amount := blockchain.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)

}
