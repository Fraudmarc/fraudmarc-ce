import { Injectable } from '@angular/core';
import { MediaChange, MediaObserver } from '@angular/flex-layout';
import { BehaviorSubject } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class MediaChangeService {

  public media =  new BehaviorSubject<string>('xs');

  constructor(mediaObserver: MediaObserver) {
    mediaObserver.media$.subscribe((change: MediaChange) => {
      this.media.next(change.mqAlias);
    });
  }
}
