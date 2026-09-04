import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';

import { ProductService } from '../../core/services/product.service';

@Component({
  selector: 'app-product-create',
  imports: [
    CommonModule,
    FormsModule,
    RouterLink
  ],
  templateUrl: './product-create.component.html',
  styleUrl: './product-create.component.css'
})
export class ProductCreateComponent {

  private productService = inject(ProductService);
  private router = inject(Router);

  code = '';
  description = '';
  stock: number | null = null;

  loading = false;
  error = '';

  createProduct(): void {
    this.error = '';

    if (!this.code.trim()) {
      this.error = 'Informe o código do produto.';
      return;
    }

    if (!this.description.trim()) {
      this.error = 'Informe a descrição do produto.';
      return;
    }

    if (this.stock === null || this.stock < 0) {
      this.error = 'Informe um estoque válido.';
      return;
    }

    this.loading = true;

    this.productService.createProduct({
      code: this.code.trim(),
      description: this.description.trim(),
      stock: this.stock
    }).subscribe({
      next: () => {
        this.loading = false;
        this.router.navigate(['/products']);
      },
      error: (error) => {
        console.error('Erro ao cadastrar produto:', error);

        this.loading = false;
        this.error = 'Não foi possível cadastrar o produto.';
      }
    });
  }
}