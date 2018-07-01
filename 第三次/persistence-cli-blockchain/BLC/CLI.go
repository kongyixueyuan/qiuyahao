package BLC

import (
	"fmt"
	"os"
	"flag"
	"log"
)

type CLI struct {
	Blockchain *Blockchain
}

func printUsage() {

	fmt.Println("Usage:")
	fmt.Println("\tcreateBlockchain -data -- 创建创世区块")
	fmt.Println("\taddBlock -data DATA -- 增加区块")
	fmt.Println("\tprintChain -- 打印区块信息")

}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) addBlock(data string) {

	if !DBExists() {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}

	blockchain := BlockchainObject()

	defer blockchain.DB.Close()

	blockchain.AddBlockToBlockchain(data)
}

func (cli *CLI) printChain() {

	if !DBExists() {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}

	blockchain := BlockchainObject()

	defer blockchain.DB.Close()

	blockchain.Printchain()
}

func (cli *CLI) createGenesisBlockchain(data string)  {

	CreateBlockchainWithGenenisBlock(data)

}

func (cli *CLI) Run() {

	addBlock := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printChain := flag.NewFlagSet("printChain", flag.ExitOnError)
	createBlockchain := flag.NewFlagSet("createBlockchain", flag.ExitOnError)

	addBlockData := addBlock.String("data", "xiaohao", "增加交易数据")

	createBlockchainData := createBlockchain.String("data", "Genesis block data", "创世区块交易数据")

	isValidArgs()

	switch os.Args[1] {
	case "addBlock":
		err := addBlock.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printChain":
		err := printChain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createBlockchain":
		err := createBlockchain.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)

	}

	if addBlock.Parsed() {

		if *addBlockData == "" {
			printUsage()
			os.Exit(1)
		}

		cli.addBlock(*addBlockData)

	}

	if printChain.Parsed() {

		cli.printChain()

	}

	if createBlockchain.Parsed() {

		if *createBlockchainData == "" {
			fmt.Println("交易数据不能为空")
			printUsage()
			os.Exit(1)
		}

		cli.createGenesisBlockchain(*createBlockchainData)

	}

}