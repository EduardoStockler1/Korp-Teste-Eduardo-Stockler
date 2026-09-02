package models

type Invoice struct {
	ID     int           `json:"id"`
	Number int           `json:"number" binding:"required,gt=0"`
	Status string        `json:"status"`
	Items  []InvoiceItem `json:"items" binding:"required,min=1,dive"`
}

type InvoiceItem struct {
	ID        int `json:"id"`
	InvoiceID int `json:"invoiceId"`
	ProductID int `json:"productId" binding:"required,gt=0"`
	Quantity  int `json:"quantity" binding:"required,gt=0"`
}
