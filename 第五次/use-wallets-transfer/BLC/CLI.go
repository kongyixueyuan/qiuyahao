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
	fmt.Println("\taddresslists -- 输出所有钱包地址")
	fmt.Println("\tcreatewallet -- 创建钱包")
	fmt.Println("\tcreateblockchain -address -- 创建创世区块")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -- 交易明细")
	fmt.Println("\tprintchain -- 输出区块信息")
	fmt.Println("\tgetbalance -address -- 输出区块信息")

}

func isValidArgs() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
}

func (cli *CLI) Run() {

	addresslistsCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)

	flagFrom := sendBlockCmd.String("from", "", "转账源地址")
	flagTo := sendBlockCmd.String("to", "", "转账目的地址")
	flagAmount := sendBlockCmd.String("amount", "", "转账金额")

	flagCreateBlockchainWithAddress := createBlockchainCmd.String("address", "", "创建创世区块的地址")
	getBalanceWithAddress := getBalanceCmd.String("address", "", "查询某个账户的余额")

	isValidArgs()

	switch os.Args[1] {
	case "send":
		err := sendBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createwallet":
		err := createWalletCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "addresslists":
		err := addresslistsCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		printUsage()
		os.Exit(1)

	}

	if sendBlockCmd.Parsed() {

		if *flagFrom == "" || *flagTo == "" || *flagAmount == "" {
			printUsage()
			os.Exit(1)
		}

		//fmt.Println(*flagAddBlockData)
		//cli.addBlock([]*Transaction{})
		//fmt.Println(*flagFrom)
		//fmt.Println(*flagTo)
		//fmt.Println(*flagAmount)
		//
		//fmt.Println(JSONToArray(*flagFrom))
		//fmt.Println(JSONToArray(*flagTo))
		//fmt.Println(JSONToArray(*flagAmount))
		from := JSONToArray(*flagFrom)
		to := JSONToArray(*flagTo)

		for index, fromAddress := range from {
			if !IsValidForAddress([]byte(fromAddress)) || !IsValidForAddress([]byte(to[index])) {
				fmt.Printf("地址无效")
				os.Exit(1)
			}
		}

		amount := JSONToArray(*flagAmount)

		cli.Send(from, to, amount)

	}

	if printChainCmd.Parsed() {

		//fmt.Println("输出所有区块的信息")
		cli.Printchain()

	}

	if createBlockchainCmd.Parsed() {

		if !IsValidForAddress([]byte(*flagCreateBlockchainWithAddress)) {
			fmt.Println("地址无效")
			printUsage()
			os.Exit(1)
		}

		cli.CreateGenesisBlockchain(*flagCreateBlockchainWithAddress)

	}

	if getBalanceCmd.Parsed() {

		if !IsValidForAddress([]byte(*getBalanceWithAddress)) {

			fmt.Println("地址无效")
			printUsage()
			os.Exit(1)
		}

		cli.GetBalance(*getBalanceWithAddress)

	}

	if createWalletCmd.Parsed() {

		cli.CreateWallet()

	}

	if addresslistsCmd.Parsed() {

		cli.AddressLists()

	}

}