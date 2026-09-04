import { Component, OnInit, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterLink } from '@angular/router';
import { Invoice, InvoiceService } from '../../core/services/invoice.service';

@Component({
  selector: 'app-invoice-list',
  imports: [CommonModule, RouterLink],
  templateUrl: './invoice-list.component.html',
  styleUrl: './invoice-list.component.css'
})
export class InvoiceListComponent implements OnInit {

  private invoiceService = inject(InvoiceService);

  invoices: Invoice[] = [];

  loading = true;
  error = '';

  ngOnInit(): void {
    this.loadInvoices();
  }

  loadInvoices(): void {
    this.loading = true;
    this.error = '';

    this.invoiceService.getInvoices().subscribe({
      next: (invoices) => {
        this.invoices = invoices;
        this.loading = false;
      },
      error: (error) => {
        console.error('Erro ao carregar notas fiscais:', error);
        this.error = 'Não foi possível carregar as notas fiscais.';
        this.loading = false;
      }
    });
  }

  getTotalItems(invoice: Invoice): number {
    return invoice.items.reduce(
      (total, item) => total + item.quantity,
      0
    );
  }

  printInvoice(invoiceId: number): void {
    this.invoiceService.printInvoice(invoiceId).subscribe({
      next: () => {
        const handleAfterPrint = () => {
          window.removeEventListener('afterprint', handleAfterPrint);
          this.loadInvoices();
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