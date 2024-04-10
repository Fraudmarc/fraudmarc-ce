import { enableProdMode } from '@angular/core';
import { platformBrowserDynamic } from '@angular/platform-browser-dynamic';

import { AppModule } from './app/app.module';
import { environment } from './environments/environment';

// import { Amplify } from 'aws-amplify';

// Configure Amplify with the environment settings
// Amplify.configure(environment.amplify);

// const currentConfig = Auth.configure();
// console.log('currentConfig:', currentConfig);

if (environment.production) {
  enableProdMode();
}

platformBrowserDynamic().bootstrapModule(AppModule);
