import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { DomainsComponent } from './components/domains/domains.component';
import { ReportComponent } from './components/report/report.component';

const routes: Routes = [
  {
    path: '',
    component: DomainsComponent,
  },
  {
    path: 'report/:domain',
    component: ReportComponent,
  }, {
    path: 'report/:domain/:start/:end',
    component: ReportComponent,
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes, { relativeLinkResolution: 'legacy' })],
  exports: [RouterModule]
})
export class AppRoutingModule { }
