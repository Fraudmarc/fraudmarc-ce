import { Injectable } from '@angular/core';
import { MediaChange, ObservableMedia } from '@angular/flex-layout';
import { BehaviorSubject } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class MediaChangeService {

  public media =  new BehaviorSubject<string>('xs');

  constructor(mediaObserver: ObservableMedia) {
    mediaObserver.subscribe((change: MediaChange) => {
        this.media.next(change.mqAlias);
    });
  }
}
