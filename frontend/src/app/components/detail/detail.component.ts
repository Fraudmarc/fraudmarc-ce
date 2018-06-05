import { DatePipe } from '@angular/common';
import { Component, Inject, OnInit, ViewChild } from '@angular/core';
import { MAT_DIALOG_DATA, MatDialogRef, MatPaginator, MatTableDataSource } from '@angular/material';
import { countries } from '../../../assets/simple-countries';
import { directCopy } from '../../app.utilities';
import { DomainDmarcDetailDataProvider, IDMARCReportDetailRecord, DashWhenEmptyStringPipe } from '../../services/dmarc.service';

@Component({
  selector: 'app-detail',
  templateUrl: './detail.component.html',
  styleUrls: ['./detail.component.scss'],
  providers: [DatePipe]
})
export class DetailComponent implements OnInit {

  constructor(
    private datePipe: DatePipe,
    public dialogRef: MatDialogRef<DetailComponent>,
    @Inject(MAT_DIALOG_DATA) public dataProvider: DomainDmarcDetailDataProvider
  ) { }

  public loading: boolean;
  public hasError: boolean;
  public errorMessage: string;
  public subtitle = 'DMARC Report';
  public dateRange: string;

  public voidDataSource: MatTableDataSource<void>; // for typing

  public detailDataProvider: any;

  public detailColumns = [
    'source_ip',
    'country',
    'message_count',
    'disposition',
    'eval_spf',
    'auth_spf',
    'auth_spf_domain',
    'eval_dkim',
    'auth_dkim_result',
    'auth_dkim_selector',
    'auth_dkim_domain',
    'po_reason',
    'po_comment'
  ];

  @ViewChild('pager') set pager(paginator: MatPaginator) { this.dataProvider.DetailDataSource.paginator = paginator; }

  ngOnInit(): void {
    this.loading = true;
    this.dataProvider.onData.subscribe(() => {
      this.loading = false;
      this.detailDataProvider = this.dataProvider.DetailDataSource;
      // console.log(this.detailDataProvider);
    });
    // this.loading = false;
    // this.detailDataProvider = this.dataProvider.DetailDataSource;
    this.dateRange = `${this.datePipe.transform(this.dataProvider.startDate, 'shortDate')} to ${this.datePipe.transform(this.dataProvider.endDate, 'shortDate')}`;
  }

  onRowClick(ev: MouseEvent, row: IDMARCReportDetailRecord, i): void {
    ev.preventDefault();
    if (ev.ctrlKey) {
      directCopy(`${row.source_ip}, ${row.country}, ${row.message_count}, ${row.disposition}, ${row.eval_spf}, ${row.auth_spf_result}, ${row.auth_spf_domain}, ${row.eval_dkim}, ${row.auth_dkim_result}, ${row.auth_dkim_selector}, ${row.auth_spf_domain}, ${row.po_reason}, ${row.po_comment}`);
    }
  }

  back(): void { this.dialogRef.close(); }

  getName(countryCode): string { return countries.names[countries.codes.indexOf(countryCode)]; }

}
