import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { delay } from 'rxjs';

import { CartItem } from '../../models/cart-item';
import { Product } from '../../models/product';

import {
  Invoice,
  InvoiceService
} from '../../core/services/invoice.service';

import { ProductService } from '../../core/services/product.service';
import { InvoiceComponent } from '../invoice/invoice.component';

@Component({
  selector: 'app-products',
  imports: [
    CommonModule,
    RouterLink,
    InvoiceComponent
  ],
  templateUrl: './products.component.html',
  styleUrl: './products.component.css'
})
export class ProductsComponent implements OnInit {

  private invoiceService = inject(InvoiceService);
  private productService = inject(ProductService);

  createdInvoice: Invoice | null = null;

  products: Product[] = [];
  cart: CartItem[] = [];

  loading = true;
  creatingInvoice = false;
  error = '';

  ngOnInit(): void {
    this.productService.getProducts().subscribe({
      next: (products) => {
        this.products = products;
        this.loading = false;
      },
      error: () => {
        this.error = 'Não foi possível carregar os produtos.';
        this.loading = false;
      }
    });
  }

  addToCart(product: Product): void {
    const existingItem = this.cart.find(
      item => item.product.id === product.id
    );

    if (existingItem) {
      if (existingItem.quantity < product.stock) {
        existingItem.quantity++;
      }

      return;
    }

    this.cart.push({
      product,
      quantity: 1
    });
  }

  increaseQuantity(item: CartItem): void {
    if (item.quantity < item.product.stock) {
      item.quantity++;
    }
  }

  decreaseQuantity(item: CartItem): void {
    if (item.quantity > 1) {
      item.quantity--;
      return;
    }

    this.cart = this.cart.filter(
      cartItem => cartItem.product.id !== item.product.id
    );
  }

  getTotalItems(): number {
    return this.cart.reduce(
      (total, item) => total + item.quantity,
      0
    );
  }

  checkout(): void {
    if (this.cart.length === 0 || this.creatingInvoice) {
      return;
    }

    this.creatingInvoice = true;
    this.error = '';

    const invoice = {
      items: this.cart.map(item => ({
        productId: item.product.id,
        quantity: item.quantity
      }))
    };

    console.log('Enviando NFSe:', invoice);

    this.invoiceService
      .createInvoice(invoice)
      .pipe(
        delay(2000)
      )
      .subscribe({
        next: (createdInvoice) => {
          console.log('NFSe criada:', createdInvoice);

          this.createdInvoice = createdInvoice;
          this.cart = [];
          this.creatingInvoice = false;
        },

        error: (error) => {
          console.error('Erro ao criar NFSe:', error);

          this.creatingInvoice = false;
          this.error = 'Não foi possível finalizar a compra.';
        }
      });
  }
}