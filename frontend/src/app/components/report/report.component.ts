import { DatePipe } from '@angular/common';
import { Component, OnInit, ViewChild } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { MatLegacyDialog as MatDialog, MatLegacyDialogConfig as MatDialogConfig } from '@angular/material/legacy-dialog';
import { MatLegacyPaginator as MatPaginator } from '@angular/material/legacy-paginator';
import { MatLegacyTableDataSource as MatTableDataSource } from '@angular/material/legacy-table';
import { IDateRange, directCopy, getlast30DayRange } from '../../app.utilities';
import { DetailComponent } from '../detail/detail.component';
import {
  IDmarcChart,
  DmarcService,
  DomainDmarcDataProvider,
  IDMARCReportSummaryRecord,
  IDMARCReportTotalRecord
} from '../../services/dmarc.service';

type ISODateString = string;

@Component({
  selector: 'app-report',
  templateUrl: './report.component.html',
  styleUrls: ['./report.component.scss'],
  providers: [DatePipe]
})
export class ReportComponent implements OnInit {

  public domainNav: string;
  public domain: string;
  public subtitle = 'DMARC Report';
  public dateRange: string;
  public errorMessage: string;

  public startDate: ISODateString;
  public endDate: ISODateString;
  public start: Date;
  public end: Date;

  public pageSize = 25;

  public summaryDataProvider: DomainDmarcDataProvider;
  public totalDataSource: MatTableDataSource<IDMARCReportTotalRecord>;
  public summaryDataSource: MatTableDataSource<IDMARCReportSummaryRecord>;
  public voidDataSource: MatTableDataSource<void>;
  public ChartDmarcResponse: IDmarcChart;

  public hasMessages = false;
  public hasChart = false;
  public chartLoading = true;
  public reportLoading = true;

  public hasDate = false;
  public hasError = false;

  public colorScheme;

  constructor(
    private router: Router,
    private route: ActivatedRoute,
    private dmarcService: DmarcService,
    private datePipe: DatePipe,
    private dialog: MatDialog,
  ) { }

  @ViewChild('pager')
  set pager(paginator: MatPaginator) {
    if (paginator) {
      this.summaryDataProvider.summaryDataSource.paginator = paginator;
    }
  }


  ngOnInit() {
    this.colorScheme = {
      domain: ['#4caf50', '#FF0000']
    };
    this.route.params.subscribe(params => {
      const { domain, start, end } = params;
      this.domain = domain;
      if (!start || !end) {
        const { startDate, endDate } = getlast30DayRange();
        this.router.navigate(
          [`/report/${this.domain}/${startDate}/${endDate}`],
          { replaceUrl: true }
        );
        return; // Seems necessary to stop initialization and take the new route.
      }

      this.domainNav = domain;

      if (domain.search(/^[0-9]+$/) !== 0) {
        this.domain = domain;
      }

      this.startDate = start;
      this.endDate = end;
      this.start = new Date(start);
      this.end = new Date(end);
      this.hasDate = true;

      this.hasError = false;
      this.dateRange = `${this.datePipe.transform(
        this.startDate
      )} to ${this.datePipe.transform(this.endDate)}`;
      this.summaryDataProvider = this.dmarcService.getSummaryDataProvider(
        domain,
        this.startDate,
        this.endDate
      );

      this.dmarcService
        .getChartData(domain, this.startDate, this.endDate)
        .subscribe(result => {
          this.chartLoading = false;
          this.ChartDmarcResponse = result;
          if (this.ChartDmarcResponse.chartdata[0].series.length) {
            this.hasChart = true;
          }
        });

      this.summaryDataProvider.onError.subscribe(err => {
        this.hasError = true;
        this.errorMessage = err;
      });

      this.totalDataSource = this.summaryDataProvider.totalDataSource;
      this.totalDataSource.connect().subscribe(totals => {
        this.reportLoading = false;
        //console.log(totals);
        if (totals.length === 0 || totals[0].message_count > 0) {
          this.hasMessages = true;
        }
        this.domain = this.summaryDataProvider.domain;
      });
      this.summaryDataSource = this.summaryDataProvider.summaryDataSource;
    });
  }

  setDateRange(range: IDateRange): void {
    this.router.navigate([
      `/report/${this.domainNav}/${range.startDate}/${range.endDate}`
    ]);
  }

  toDomains(): void {
    this.router.navigate([``]);
  }

  backAction(): void {
    this.toDomains();
  }

  onRowClick(ev: MouseEvent, row: IDMARCReportSummaryRecord, i) {
    ev.preventDefault();
    if (ev.ctrlKey) {
      directCopy(
        `${row.source}, ${row.total_count}, ${row.fully_aligned_count}, ${
          row.spf_aligned_count
        }, ${row.dkim_aligned_count}`
      );
    } else {
      const detailProvider = this.dmarcService.getDetailDataProvider(
        this.domainNav,
        this.domain,
        this.startDate,
        this.endDate,
        row.source,
        row.source_type
      );
      const config: MatDialogConfig = {
        data: detailProvider,
        height: 'calc(100vh - 10px)',
        width: 'calc(100vw - 30px)',
        maxWidth: '100vw',
        panelClass: 'sourceDetail',
        autoFocus: false
      };
      const dialogRef = this.dialog.open(DetailComponent, config);
      dialogRef
        .afterClosed()
        .subscribe(result => console.log('The dialog was closed'));
    }
  }

  toTop(target: HTMLDivElement) {
    target.children.item(0).scrollIntoView();
  }
}
