package models

type ExecutionRecord struct {
	ID            int     
	BuyOrderID    int     
	SellOrderID   int     
	Symbol        string    
	Price         float64   
	Quantity      int
}
