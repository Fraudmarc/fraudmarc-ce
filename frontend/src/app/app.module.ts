import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { FlexLayoutModule } from '@angular/flex-layout';
import { MatButtonModule, MatCardModule, MatCheckboxModule, MatChipsModule, MatDatepickerModule, MatDialogModule, MatExpansionModule, MatFormFieldModule, MatGridListModule, MatIconModule, MatInputModule, MatListModule, MatMenuModule, MatNativeDateModule, MatPaginatorModule, MatProgressBarModule, MatProgressSpinnerModule, MatRadioModule, MatSelectModule, MatSidenavModule, MatSlideToggleModule, MatSliderModule, MatSnackBarModule, MatSortModule, MatTableModule, MatTabsModule, MatToolbarModule, MatTooltipModule } from '@angular/material';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';

import { NgxChartsModule } from '@swimlane/ngx-charts';
import { NgxDatatableModule } from '@swimlane/ngx-datatable';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import 'hammerjs';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { DomainsComponent } from './components/domains/domains.component';
import { HeaderComponent } from './components/header/header.component';
import { FooterComponent } from './components/footer/footer.component';

import { MediaChangeService } from './services/media-change.service';
import { DashWhenEmptyStringPipe, DmarcService} from './services/dmarc.service';

import { PageHeaderComponent } from './components/page-header/page-header.component';
import { SearchboxComponent } from './components/searchbox/searchbox.component';
import { DateRangeComponent } from './components/date-range/date-range.component';
import { LineChartComponent } from './components/line-chart/line-chart.component';
import { ReportComponent } from './components/report/report.component';
import { DetailComponent } from './components/detail/detail.component';
import { ProgressPanelComponent } from './components/progress-panel/progress-panel.component';

@NgModule({
  declarations: [
    DashWhenEmptyStringPipe,
    AppComponent,
    DomainsComponent,
    HeaderComponent,
    FooterComponent,
    PageHeaderComponent,
    SearchboxComponent,
    DateRangeComponent,
    LineChartComponent,
    ReportComponent,
    DetailComponent,
    ProgressPanelComponent,
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    FlexLayoutModule,
    MatButtonModule,
    MatCardModule,
    MatCheckboxModule,
    MatChipsModule,
    MatDialogModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatExpansionModule,
    MatFormFieldModule,
    MatGridListModule,
    MatIconModule,
    MatInputModule,
    MatListModule,
    MatMenuModule,
    MatPaginatorModule,
    MatProgressBarModule,
    MatProgressSpinnerModule,
    MatRadioModule,
    MatSelectModule,
    MatSidenavModule,
    MatSlideToggleModule,
    MatSliderModule,
    MatSnackBarModule,
    MatTabsModule,
    MatTableModule,
    MatSortModule,
    MatToolbarModule,
    MatTooltipModule,
    FormsModule,
    ReactiveFormsModule,
    NgxChartsModule,
    NgxDatatableModule,
    BrowserAnimationsModule 
  ],
  entryComponents: [
    DetailComponent
  ],
  providers: [
    MediaChangeService,
    DmarcService
  ],
  bootstrap: [AppComponent]
})
export class AppModule { }
