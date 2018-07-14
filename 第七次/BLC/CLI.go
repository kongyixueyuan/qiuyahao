package BLC

import (
	"fmt"
	"os"
	"flag"
	"log"
)

type QYH_CLI struct {}


func QYH_printUsage()  {

	fmt.Println("Usage:")

	fmt.Println("\taddresslists -- 输出所有钱包地址.")
	fmt.Println("\tcreatewallet -- 创建钱包.")
	fmt.Println("\tcreateblockchain -address -- 交易数据.")
	fmt.Println("\tsend -from FROM -to TO -amount AMOUNT -mine -- 交易明细.")
	fmt.Println("\tprintchain -- 输出区块信息.")
	fmt.Println("\tgetbalance -address -- 输出区块信息.")
	fmt.Println("\tresetUTXO -- 重置.")
	fmt.Println("\tstartnode -miner ADDRESS -- 启动节点服务器，并且指定挖矿奖励的地址.")

}

func QYH_isValidArgs()  {
	if len(os.Args) < 2 {
		QYH_printUsage()
		os.Exit(1)
	}
}



func (cli *QYH_CLI) QYH_Run()  {

	QYH_isValidArgs()

	//获取节点ID

	// 设置ID
	// export NODE_ID=8888
	// 读取
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID env. var is not set!\n")
		os.Exit(1)
	}

	fmt.Printf("NODE_ID:%s\n",nodeID)

	resetUTXOCMD := flag.NewFlagSet("resetUTXO",flag.ExitOnError)
	addresslistsCmd := flag.NewFlagSet("addresslists",flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet",flag.ExitOnError)
	sendBlockCmd := flag.NewFlagSet("send",flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain",flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain",flag.ExitOnError)
	getbalanceCmd := flag.NewFlagSet("getbalance",flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode",flag.ExitOnError)


	flagFrom := sendBlockCmd.String("from","","转账源地址......")
	flagTo := sendBlockCmd.String("to","","转账目的地地址......")
	flagAmount := sendBlockCmd.String("amount","","转账金额......")
	flagMine := sendBlockCmd.Bool("mine",false,"是否在当前节点中立即验证....")


	flagMiner := startNodeCmd.String("miner","","定义挖矿奖励的地址......")

	flagCreateBlockchainWithAddress := createBlockchainCmd.String("address","","创建创世区块的地址")
	getbalanceWithAdress := getbalanceCmd.String("address","","要查询某一个账号的余额.......")



	switch os.Args[1] {
		case "send":
			err := sendBlockCmd.Parse(os.Args[2:])
			if err != nil {
				log.Panic(err)
			}
		case "startnode":
			err := startNodeCmd.Parse(os.Args[2:])
			if err != nil {
				log.Panic(err)
			}
		case "resetUTXO":
			err := resetUTXOCMD.Parse(os.Args[2:])
			if err != nil {
				log.Panic(err)
			}
		case "addresslists":
			err := addresslistsCmd.Parse(os.Args[2:])
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
			err := getbalanceCmd.Parse(os.Args[2:])
			if err != nil {
				log.Panic(err)
			}
		case "createwallet":
			err := createWalletCmd.Parse(os.Args[2:])
			if err != nil {
				log.Panic(err)
			}
		default:
			QYH_printUsage()
			os.Exit(1)
	}

	if sendBlockCmd.Parsed() {
		if *flagFrom == "" || *flagTo == "" || *flagAmount == ""{
			QYH_printUsage()
			os.Exit(1)
		}

		from := JSONToArray(*flagFrom)
		to := JSONToArray(*flagTo)

		for index,fromAdress := range from {
			if QYH_IsValidForAdress([]byte(fromAdress)) == false || QYH_IsValidForAdress([]byte(to[index])) == false {
				fmt.Printf("地址无效......")
				QYH_printUsage()
				os.Exit(1)
			}
		}

		amount := JSONToArray(*flagAmount)
		cli.QYH_send(from,to,amount,nodeID,*flagMine)
	}

	if printChainCmd.Parsed() {

		//fmt.Println("输出所有区块的数据........")
		cli.QYH_printchain(nodeID)
	}

	if resetUTXOCMD.Parsed() {

		fmt.Println("重置UTXO表单......")
		cli.QYH_resetUTXOSet(nodeID)
	}

	if addresslistsCmd.Parsed() {

		//fmt.Println("输出所有区块的数据........")
		cli.QYH_addressLists(nodeID)
	}


	if createWalletCmd.Parsed() {
		// 创建钱包
		cli.QYH_createWallet(nodeID)
	}

	if createBlockchainCmd.Parsed() {

		if QYH_IsValidForAdress([]byte(*flagCreateBlockchainWithAddress)) == false {
			fmt.Println("地址无效....")
			QYH_printUsage()
			os.Exit(1)
		}


		cli.QYH_createGenesisBlockchain(*flagCreateBlockchainWithAddress,nodeID)
	}

	if getbalanceCmd.Parsed() {

		if QYH_IsValidForAdress([]byte(*getbalanceWithAdress)) == false {
			fmt.Println("地址无效....")
			QYH_printUsage()
			os.Exit(1)
		}

		cli.QYH_getBalance(*getbalanceWithAdress,nodeID)
	}

	if startNodeCmd.Parsed() {


		cli.QYH_startNode(nodeID,*flagMiner)
	}

}