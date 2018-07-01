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
	fmt.Println("\tcreateBlockchain -address -- 创建创世区块")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -- 发起交易")
	fmt.Println("\tprintChain -- 打印区块信息")
	fmt.Println("\tgetBalance -address -- 查询余额")

}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {

	sendBlock := flag.NewFlagSet("send", flag.ExitOnError)
	printChain := flag.NewFlagSet("printChain", flag.ExitOnError)
	createBlockchain := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	getBalance := flag.NewFlagSet("getBalance", flag.ExitOnError)

	flagFrom := sendBlock.String("from", "", "转账源地址")
	flagTo := sendBlock.String("to", "", "转账目的地址")
	flagAmount := sendBlock.String("amount", "", "转账金额")

	flagCreateBlockchainWithAddress := createBlockchain.String("address", "", "创建创世区块的地址")
	getBalanceWithAddress := getBalance.String("address", "", "查询某个账户的余额")

	isValidArgs()

	switch os.Args[1] {
	case "send":
		err := sendBlock.Parse(os.Args[2:])
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
	case "getBalance":
		err := getBalance.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)

	}

	if sendBlock.Parsed() {

		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}

		from := JSONToArray(*flagFrom)
		to := JSONToArray(*flagTo)
		amount := JSONToArray(*flagAmount)

		cli.Send(from, to, amount)

	}

	if printChain.Parsed() {

		cli.Printchain()

	}

	if createBlockchain.Parsed() {

		if *flagCreateBlockchainWithAddress == "" {
			fmt.Println("地址不能为空")
			printUsage()
			os.Exit(1)
		}

		cli.CreateGenesisBlockchain(*flagCreateBlockchainWithAddress)

	}

	if getBalance.Parsed() {

		if *getBalanceWithAddress == "" {
			fmt.Println("地址不能为空")
			printUsage()
			os.Exit(1)
		}

		cli.GetBalance(*getBalanceWithAddress)

	}

}

// 创建创世区块
func (cli *CLI) CreateGenesisBlockchain(address string)  {

	blockchain := CreateBlockchainWithGenenisBlock(address)
	defer blockchain.DB.Close()

}

// 查询余额
func (cli *CLI) GetBalance(address string)  {

	blockchain := BlockchainObject()
	defer blockchain.DB.Close()

	amount := blockchain.GetBalance(address)

	fmt.Printf("%s一共有%d个Token\n", address, amount)

}

func (cli *CLI) Printchain() {

	if DBExists() == false {
		fmt.Println("数据库不存在")
		os.Exit(1)
	}

	blockchain := BlockchainObject()

	defer blockchain.DB.Close()

	blockchain.Printchain()
}

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