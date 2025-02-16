package models

type InfoResponse struct {
	Coins       int                    `json:"coins"`
	Inventory   []*InventoryItem       `json:"inventory"`
	CoinHistory CoinTransactionHistory `json:"coinHistory"`
}

type InventoryItem struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinTransactionHistory struct {
	Received []CoinReceived `json:"received"`
	Sent     []CoinSent     `json:"sent"`
}

type CoinReceived struct {
	FromUser string `json:"fromUser"`
	Amount   int    `json:"amount"`
}

type CoinSent struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}
