package apitypes

import "github.com/sinsinpurin/gomyblockchain"

type GetTransactionsResponseType struct {
	Transactions []gomyblockchain.Transaction `json:"transactions"`
	Length       int                          `json:"length"`
}

type PostTransactionsRequestType struct {
	RecipientAddress string `json:"RecipientAddress"`
	SenderAddress    string `json:"SenderAddress"`
	Value            uint64 `json:"Value"`
	SenderPublicKey  string `json:"SenderPublicKey"`
	Signature        string `json:"Signature"`
}

type PostMiningRequestType struct {
	MinerAddress string `json:"MinerAddress"`
}

type GetAmountRequestType struct {
	Address string `json:"Address"`
}
type GetAmountResponseType struct {
	Amount uint64 `json:"amount"`
}
