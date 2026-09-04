import { Injectable, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';

import { Product } from '../../models/product';

@Injectable({
  providedIn: 'root'
})
export class ProductService {

  private http = inject(HttpClient);

  private readonly apiUrl = 'http://localhost:8081';

  getProducts(): Observable<Product[]> {
    return this.http.get<Product[]>(`${this.apiUrl}/products`);
  }

  createProduct(product: {
    code: string;
    description: string;
    stock: number;
  }): Observable<Product> {
    return this.http.post<Product>(
      `${this.apiUrl}/products`,
      product
    );
  }

  decreaseStock(
    productId: number,
    quantity: number
  ): Observable<Product> {
    return this.http.post<Product>(
      `${this.apiUrl}/products/${productId}/decrease-stock`,
      { quantity }
    );
  }
}