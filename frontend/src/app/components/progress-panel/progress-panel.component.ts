import { Component } from '@angular/core';

@Component({
  selector: 'fm-progress-panel',
  template: `
    <div fxLayout="column"
         fxLayoutAlign="center center"
         [ngStyle]="{'height':'300px'}">
      <mat-progress-spinner mode="indeterminate"
                            strokeWidth="2">
      </mat-progress-spinner>
    </div>
  `,
  styleUrls: []
})
export class ProgressPanelComponent { }