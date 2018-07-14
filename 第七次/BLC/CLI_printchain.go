package BLC


func (cli *QYH_CLI) QYH_printchain(nodeID string)  {

	blockchain := QYH_BlockchainObject(nodeID)

	defer blockchain.QYH_DB.Close()

	blockchain.QYH_Printchain()

}