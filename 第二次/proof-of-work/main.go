package main

import (
	"qiuyahao/第二次/proof-of-work/BLC"
	"fmt"
)

func main() {

	block := BLC.NewBlock("Test", 1, []byte{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0})

	fmt.Printf("%d\n", block.Nonce)
	fmt.Printf("%x\n", block.Hash)

	proofOfWork := BLC.NewProofOfWork(block)

	fmt.Printf("%v", proofOfWork.IsValid())
}
