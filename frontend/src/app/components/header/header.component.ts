import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from '../../../environments/environment';
import { MediaChangeService } from '../../services/media-change.service';

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.scss']
})
export class HeaderComponent implements OnInit {

  public breakpoint: string;
  public isProduction: boolean;

  constructor(
    private router: Router,
    private mediaChange: MediaChangeService
  ) { }

  ngOnInit() {
    this.mediaChange.media.subscribe((breakpoint) => {
      this.breakpoint = breakpoint;
    });
  }

  goHome(event: MouseEvent) {
    if (event.altKey) {
      this.router.navigate(['/jobs']);
    } else {
      this.router.navigate(['/domains']);
    }
  }

}
