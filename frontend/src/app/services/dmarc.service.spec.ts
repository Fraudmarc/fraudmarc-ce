import { TestBed, inject } from '@angular/core/testing';

import { DmarcService } from './dmarc.service';

describe('DmarcService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
    providers: [DmarcService],
    teardown: { destroyAfterEach: false }
});
  });

  it('should be created', inject([DmarcService], (service: DmarcService) => {
    expect(service).toBeTruthy();
  }));
});
