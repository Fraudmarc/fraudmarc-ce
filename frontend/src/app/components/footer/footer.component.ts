import { Component, OnInit } from '@angular/core';
import {
  IDmarcChart,
  DmarcService
} from '../../services/dmarc.service';

@Component({
  selector: 'app-footer',
  templateUrl: './footer.component.html',
  styleUrls: ['./footer.component.scss']
})
export class FooterComponent implements OnInit {

  ChartDmarcResponse: IDmarcChart;
  testDomain: string;
  startDate: string;
  endDate: string;

  domainList: any;


  constructor(private dmarcService: DmarcService) { }

  ngOnInit() {
  }

  testBackendCall() {
    this.dmarcService.getChartData(this.testDomain, this.startDate, this.endDate).subscribe(res => {
      this.ChartDmarcResponse = res;
    })
  }
  // ${environment.APIBASEURL}/domains

  testDomainList() { 
    this.dmarcService.getDomainList().subscribe(res =>{
      console.log(res);
      this.domainList = res;
    }

    )
  }

}
