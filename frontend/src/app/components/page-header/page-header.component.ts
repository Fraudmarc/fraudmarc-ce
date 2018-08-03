import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-page-header',
  templateUrl: './page-header.component.html',
  styleUrls: ['./page-header.component.scss']
})
export class PageHeaderComponent implements OnInit {

  @Input() title: string;
  @Input() subtitle: string;
  @Input() nonav: boolean;
  @Input() action: string;
  @Output() onaction = new EventEmitter<void>();

  public showFeature: boolean;
  public featureList: string[];
  constructor(private router: Router) {}

  ngOnInit() {
    this.showFeature = false;
    this.featureList = [
      'SPF CompressionSM',
      'SPF Editor',
      'DMARC Editor ',
      'STS Editor',
      'Senders Reports',
      'SPF History',
      'Failure Reports',
      'IP Intelligence: identifies major Sender by IP',
      'IP Reputation: identifies blacklisted IPs',
      'Risk Analysis',
      'Auto Reject',
      'Sender suggestions & one-click policy editing- within Senders Reports',
      'Instant setup Integration- for GoDaddy and Cloudflare domains',
      'Domain Dashboard features: quick view information, List View, sort, filters',
      'SPF Sparkline',
      'DKIM Manager'
    ];
  }

 homeAction () {
   if (!this.nonav) {
     this.router.navigate(['']);
   }
 }

 show() {
   this.showFeature = true;
 }
 hide() {
   this.showFeature = false;
 }

}
