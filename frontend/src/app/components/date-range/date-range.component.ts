import { Component, EventEmitter, Input, OnInit, Output } from '@angular/core';
import { IDateRange, toMidnight } from '../../app.utilities';

type ISODateString = string;


@Component({
  selector: 'fm-date-range',
  templateUrl: './date-range.component.html',
  styleUrls: ['./date-range.component.scss']
})
export class DateRangeComponent implements OnInit {

  @Input() start: Date;
  @Input() end: Date;
  @Input() inset: boolean;
  @Output() onDateRange = new EventEmitter<IDateRange>();

  constructor() { }

  ngOnInit() {

  }

  setRange() {
    // check range
    // start < end
    // end <= today
    if (this.start.valueOf() > this.end.valueOf()) {
      return;
    }
    this.onDateRange.emit({
      startDate: toMidnight(this.start, true),
      endDate: toMidnight(this.end)
    });
  }


  maxEnd() { return new Date(toMidnight(new Date(), true)); }
  maxStart() { return new Date( toMidnight(this.end, true)); }
  minStart() { return new Date( toMidnight(this.start)); }

}
