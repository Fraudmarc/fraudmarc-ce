import { Injectable } from '@angular/core';
import { MediaChange, MediaObserver } from '@angular/flex-layout';
import { BehaviorSubject } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class MediaChangeService {

  public media = new BehaviorSubject<string>('xs');

  constructor(mediaObserver: MediaObserver) {
    mediaObserver.asObservable().subscribe((changes: MediaChange[]) => {
        // Assuming you want to handle the last change if multiple changes are emitted at once
        const lastChange = changes[changes.length - 1];
        this.media.next(lastChange.mqAlias);
    });
  }
}