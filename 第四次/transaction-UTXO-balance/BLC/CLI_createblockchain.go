package BLC

// 创建创世区块
func (cli *CLI) CreateGenesisBlockchain(address string)  {

	blockchain := CreateBlockchainWithGenenisBlock(address)
	defer blockchain.DB.Close()

}