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

  constructor(private router: Router) {}

  ngOnInit(): void { }

 homeAction () {
   if (!this.nonav) {
     this.router.navigate(['']);
   }
 }

}
