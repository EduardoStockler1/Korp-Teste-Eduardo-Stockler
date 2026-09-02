package models

type Invoice struct {
	ID     int           `json:"id"`
	Number int           `json:"number"`
	Status string        `json:"status"` // pensei em criar usando boolean, mas talvez seja melhor usar string para melhor leitura e também porque seria uma microotimização desnecessária, já que o banco de dados vai armazenar como string mesmo
	Items  []InvoiceItem `json:"items"`
}

type InvoiceItem struct {
	ID        int `json:"id"`
	InvoiceID int `json:"invoiceId"`
	ProductID int `json:"productId"` // não há foreign key para o productId, pois o produto pode ser excluído do estoque, mas ainda assim queremos manter o registro da nota fiscal
	Quantity  int `json:"quantity"`
}
