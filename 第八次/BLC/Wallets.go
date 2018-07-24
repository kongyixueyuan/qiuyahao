package BLC

import (
	"os"
	"io/ioutil"
	"log"
	"encoding/gob"
	"crypto/elliptic"
	"bytes"
	"fmt"
)

const walletFile  = "wallet_%s.dat"

type QYH_Wallets struct {
	QYH_Wallets map[string]*QYH_Wallet
}

// 生成新的钱包
// 从数据库中读取，如果不存在
func QYH_NewWallets(nodeID string)(*QYH_Wallets,error)  {
	wallets := QYH_Wallets{}
	wallets.QYH_Wallets = make(map[string]*QYH_Wallet)

	err := wallets.QYH_LoadFromFile(nodeID)

	return &wallets,err
}
// 生成新的钱包地址列表
func (ws *QYH_Wallets) QYH_NewWallet() *QYH_Wallet {
	wallet := QYH_NewWallet()
	address := wallet.QYH_GetAddress()
	ws.QYH_Wallets[string(address)] = wallet
	return wallet
}
// 获取钱包地址
func (ws *QYH_Wallets) QYH_GetAddresses()[]string  {
	var addresses []string
	for address := range ws.QYH_Wallets{
		addresses = append(addresses,address)
	}
	return addresses
}

// 根据地址获取钱包的详细信息
func (ws QYH_Wallets) QYH_GetWallet(address string) QYH_Wallet {
	return *ws.QYH_Wallets[address]
}

// 从数据库中读取钱包列表
func (ws *QYH_Wallets) QYH_LoadFromFile(nodeID string) error  {
	 walletFile := fmt.Sprintf(walletFile, nodeID)
	 if _,err := os.Stat(walletFile) ; os.IsNotExist(err){
	 	return err
	 }

	 fileContent ,err := ioutil.ReadFile(walletFile)
	 if err !=nil{
	 	log.Panic(err)
	 }

	 var wallets QYH_Wallets
	 gob.Register(elliptic.P256())
	 decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	 err = decoder.Decode(&wallets)
	 if err !=nil{
	 	log.Panic(err)
	 }

	 ws.QYH_Wallets = wallets.QYH_Wallets

	 return nil
}

// 将钱包存到数据库中
func (ws *QYH_Wallets) QYH_SaveToFile(nodeID string)  {
	walletFile := fmt.Sprintf(walletFile, nodeID)
	var content bytes.Buffer

	gob.Register(elliptic.P256())
	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err !=nil{
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile,content.Bytes(),0644)
	if err !=nil{
		log.Panic(err)
	}
}
// 打印所有钱包的余额
func (ws *QYH_Wallets) QYH_GetBalanceAll(nodeID string) map[string]int {
	addresses := ws.QYH_GetAddresses()
	bc := QYH_NewBlockchain(nodeID)
	defer bc.QYH_db.Close()
	UTXOSet := QYH_UTXOSet{bc}

	result := make(map[string]int)
	for _,address := range addresses{
		if !QYH_ValidateAddress(address) {
			result[address] = -1
		}
		balance := UTXOSet.QYH_GetBalance(address)
		result[address] = balance
	}
	return result
}