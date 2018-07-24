package BLC

import (
	"fmt"
	"os"
	"flag"
	"log"
	"strings"
)

type QYH_CLI struct{}

// 打印使用说明
func (cli *QYH_CLI) QYH_printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  createwallet - 创建钱包")
	fmt.Println("  addresslists - 打印钱包地址")
	fmt.Println("  createblockchain -address ADDRESS - 创建区块链")
	fmt.Println("  getbalance -address ADDRESS - 获取地址的余额")
	fmt.Println("  getbalanceall - 打印所有钱包地址的余额")
	fmt.Println("  printchain - 打印区块链中的所有区块数据")
	fmt.Println("  send -from FROM -to TO -amount AMOUNT -mine 转账,-mine 为是否立刻挖矿")
	fmt.Println("  reindexutxo - 重建UTXO set")
	fmt.Println("  printutxo - 打印所有的UTXO set")
	fmt.Println("  startnode -miner ADDRESS - 启动节点服务器，如果指定挖矿地址，则为挖矿服务器")
}
// 验证参数
func (cli *QYH_CLI) QYH_validateArgs()  {
	if len(os.Args) <2 {
		cli.QYH_printUsage()
		os.Exit(1)
	}
}

func (cli QYH_CLI) QYH_Run()  {

	cli.QYH_validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		fmt.Printf("NODE_ID未设置,export NODE_ID=3000\n")
		os.Exit(1)
	}

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createwallet", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("addresslists", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printUTXOCmd := flag.NewFlagSet("printutxo", flag.ExitOnError)
	getBalanceAllCmd := flag.NewFlagSet("getbalanceall", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "查询余额地址")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "创建创世区块地址")
	sendFrom := sendCmd.String("from", "", "转出账地址")
	sendTo := sendCmd.String("to", "", "转到账地址")
	sendAmount := sendCmd.String("amount", "", "转账金额")
	sendMine := sendCmd.Bool("mine", false, "立即挖矿")
	startNodeMiner := startNodeCmd.String("miner", "", "是否为挖矿节点")

	var err error
	switch os.Args[1] {
	case "getbalance":
		err = getBalanceCmd.Parse(os.Args[2:])
	case "createblockchain":
		err = createBlockchainCmd.Parse(os.Args[2:])
	case "createwallet":
		err = createWalletCmd.Parse(os.Args[2:])
	case "addresslists":
		err = listAddressesCmd.Parse(os.Args[2:])
	case "printchain":
		err = printChainCmd.Parse(os.Args[2:])
	case "reindexutxo":
		err = reindexUTXOCmd.Parse(os.Args[2:])
	case "send":
		err = sendCmd.Parse(os.Args[2:])
	case "printutxo":
		err = printUTXOCmd.Parse(os.Args[2:])
	case "getbalanceall":
		err = getBalanceAllCmd.Parse(os.Args[2:])
	case "startnode":
		err = startNodeCmd.Parse(os.Args[2:])
	default:
		cli.QYH_printUsage()
		os.Exit(1)
	}

	if err !=nil {
		log.Panic(err)
	}

	if createWalletCmd.Parsed() {
		cli.QYH_createWallet(nodeID)
	}

	if listAddressesCmd.Parsed() {
		cli.QYH_listAddrsss(nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.QYH_createblockchain(*createBlockchainAddress,nodeID)
	}

	if printChainCmd.Parsed() {
		cli.QYH_printchain(nodeID)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount == "" {
			sendCmd.Usage()
			os.Exit(1)
		}

		// 检查参数，有效性
		// 判断是否是json格式
		if strings.Contains(*sendFrom,"[") && strings.Contains(*sendFrom,"]") {
			from := JSONToArray(*sendFrom)
			to := JSONToArray(*sendTo)

			for index, fromAdress := range from {
				if !QYH_ValidateAddress(fromAdress) || !QYH_ValidateAddress(to[index]) {
					fmt.Println("地址无效。。")
					os.Exit(1)
				}
			}
			amount := JSONToArray(*sendAmount)
			cli.QYH_send(from, to, amount, nodeID, *sendMine)
		}else{
			from := []string{*sendFrom}
			to := []string{*sendTo}
			amount := []string{*sendAmount}

			cli.QYH_send(from, to, amount, nodeID, *sendMine)
		}
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.QYH_getBalance(*getBalanceAddress,nodeID)
	}
	if reindexUTXOCmd.Parsed() {
		cli.QYH_reindexUTXO(nodeID)
	}

	if printUTXOCmd.Parsed(){
		cli.QYH_printutxo(nodeID)
	}
	if getBalanceAllCmd.Parsed(){
		cli.QYH_getBalanceAll(nodeID)
	}
	if startNodeCmd.Parsed(){
		nodeID := os.Getenv("NODE_ID")
		if nodeID == "" {
			startNodeCmd.Usage()
			os.Exit(1)
		}
		cli.QYH_startNode(nodeID, *startNodeMiner)
	}

}