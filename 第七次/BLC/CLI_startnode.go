package BLC

import (
	"fmt"
	"os"
)

func (cli *QYH_CLI) QYH_startNode(nodeID string,minerAdd string)  {

	// 启动服务器

	if minerAdd == "" || QYH_IsValidForAdress([]byte(minerAdd))  {
		//  启动服务器
		fmt.Printf("启动服务器:localhost:%s\n",nodeID)
		QYH_startServer(nodeID,minerAdd)

	} else {

		fmt.Println("指定的地址无效....")
		os.Exit(0)
	}

}