import { Routes } from '@angular/router';

export const routes: Routes = [

  {
    path: '',
    redirectTo: 'products',
    pathMatch: 'full'
  },

  {
    path: 'products',
    loadComponent: () =>
      import('./pages/products/products.component')
        .then(m => m.ProductsComponent)
  },

  {
    path: 'products/new',
    loadComponent: () =>
      import('./pages/product-create/product-create.component')
        .then(m => m.ProductCreateComponent)
  },

  {
    path: 'invoice',
    loadComponent: () =>
      import('./pages/invoice-list/invoice-list.component')
        .then(m => m.InvoiceListComponent)
  }

];  