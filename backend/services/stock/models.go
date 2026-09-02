package main

type Product struct {
	ID          int    `json:"id"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Stock       int    `json:"stock"`
}
