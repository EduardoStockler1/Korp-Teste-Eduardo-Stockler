import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, map } from 'rxjs';

export interface InvoiceItemRequest {
  productId: number;
  quantity: number;
}

export interface CreateInvoiceRequest {
  items: InvoiceItemRequest[];
}

export interface Invoice {
  id: number;
  number: number;
  status: string;
  items: {
    id: number;
    invoiceId: number;
    productId: number;
    productDescription: string;
    quantity: number;
  }[];
}

interface InvoicesResponse {
  data: Invoice[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

@Injectable({
  providedIn: 'root'
})
export class InvoiceService {

  private http = inject(HttpClient);

  private readonly apiUrl = 'http://localhost:8082';

  createInvoice(
    invoice: CreateInvoiceRequest
  ): Observable<Invoice> {
    return this.http.post<Invoice>(
      `${this.apiUrl}/invoices`,
      invoice
    );
  }

  getInvoices(): Observable<Invoice[]> {
    return this.http
      .get<InvoicesResponse>(
        `${this.apiUrl}/invoices?limit=100`
      )
      .pipe(
        map(response => response.data)
      );
  }

  printInvoice(
    invoiceId: number
  ): Observable<{ message: string; status: string }> {
    return this.http.post<{ message: string; status: string }>(
      `${this.apiUrl}/invoices/${invoiceId}/print`,
      {}
    );
  }
}
