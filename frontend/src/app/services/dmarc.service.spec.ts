import { TestBed, inject } from '@angular/core/testing';

import { DmarcService } from './dmarc.service';

describe('DmarcService', () => {
  beforeEach(() => {
    TestBed.configureTestingModule({
      providers: [DmarcService]
    });
  });

  it('should be created', inject([DmarcService], (service: DmarcService) => {
    expect(service).toBeTruthy();
  }));
});
