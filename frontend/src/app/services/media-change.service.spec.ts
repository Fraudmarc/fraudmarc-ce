import { TestBed, inject } from '@angular/core/testing';

import { MediaChangeService } from './media-change.service';

describe('MediaChangeService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
    providers: [MediaChangeService],
    teardown: { destroyAfterEach: false }
});
  });

  it('should be created', inject([MediaChangeService], (service: MediaChangeService) => {
    expect(service).toBeTruthy();
  }));
});
