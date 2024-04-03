import { HttpClient } from '@angular/common/http';
import { Component, DoCheck, ElementRef, OnInit, ViewChild } from '@angular/core';
import { UntypedFormControl } from '@angular/forms';
import { MatDialog } from '@angular/material/dialog';
import { ActivatedRoute, Router } from '@angular/router';
import { BehaviorSubject } from 'rxjs';
import { environment } from '../../../environments/environment';
import { MediaChangeService } from '../../services/media-change.service';
import { DmarcService } from '../../services/dmarc.service';



@Component({
    selector: 'app-domains',
    templateUrl: './domains.component.html',
    styleUrls: ['./domains.component.scss']
})

export class DomainsComponent implements OnInit, DoCheck {
    @ViewChild('filter') filter: ElementRef;
    domainListOrigin: String[] = [];
    domainList: String[] = [];


    public domainInput = new UntypedFormControl();
    public columnCount$ = new BehaviorSubject<number>(1);

    public user: string;

    public domain_title = 'domain-title';
    public gridRowHeight = '40px';
    public filterText = '';

    public volumeStyle = {};


    public loading: boolean;
    public domainNameStyleOverride;

    constructor(
        private route: ActivatedRoute,
        private router: Router,
        public dialog: MatDialog,
        public media: MediaChangeService,
        private http: HttpClient,
        private dmarcService: DmarcService
    ) { }

    ngDoCheck() {
        this.columnCount$.next({ xl: 3, lg: 3, md: 2, sm: 2, xs: 1 }[this.media.media.value]);
    }

    ngOnInit() {
        this.loading = true;
        this.dmarcService.getDomainList()
            .subscribe((result: any) => {
                this.domainListOrigin = result;
                this.domainList = this.domainListOrigin;
                this.loading = false;
                const storedSearch = window.sessionStorage.getItem('filterString');
                if (storedSearch.length > 0) {
                    // Note: if this is two way bound it will trigger a filter
                    this.filterText = storedSearch;
                }
            }, err => {
                console.log(err);
            }, () => {
              this.loading = false;
            });
    }

    onFilter(value) {
        if (value) {
            window.sessionStorage.setItem('filterString', this.filterText);
        } else {
            window.sessionStorage.removeItem('filterString');
        }
        this.doFilter();
    }

    doFilter() {
        this.domainList = [];
        this.domainListOrigin.forEach(domain => {
            if (domain.indexOf(this.filterText.toLowerCase()) > -1) {
                this.domainList.push(domain);
            }
        });
    }

    goReport(domain: string) {
        this.router.navigate(['report', domain]);
        window.sessionStorage.setItem('filterString', this.filterText);
    }
}