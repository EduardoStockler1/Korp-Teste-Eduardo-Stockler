import { Component, Input, inject } from '@angular/core';
import { CommonModule } from '@angular/common';

import {
  Invoice,
  InvoiceService
} from '../../core/services/invoice.service';

import { Router } from '@angular/router';

@Component({
  selector: 'app-invoice',
  imports: [CommonModule],
  templateUrl: './invoice.component.html',
  styleUrl: './invoice.component.css'
})
export class InvoiceComponent {

  @Input() invoice: Invoice | null = null;

  private invoiceService = inject(InvoiceService);
  private router = inject(Router);

  getTotalItems(): number {
    if (!this.invoice) {
      return 0;
    }

    return this.invoice.items.reduce(
      (total, item) => total + item.quantity,
      0
    );
  }

  print(): void {
    if (!this.invoice) {
      return;
    }

    this.invoiceService.printInvoice(this.invoice.id).subscribe({
      next: () => {
        const handleAfterPrint = () => {
          window.removeEventListener('afterprint', handleAfterPrint);
          window.location.reload();
        };

        window.addEventListener('afterprint', handleAfterPrint);

        window.print();
      },
        error: (error) => {
          console.error('Erro ao imprimir NFSe:', error);
          alert('Não foi possível imprimir a nota fiscal.');
      }
    });
  }

}