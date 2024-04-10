#!/usr/bin/env node
import 'source-map-support/register';
import * as cdk from 'aws-cdk-lib';
import { DomainCdkStack } from '../lib/cdk-domain';
import { AppCdkStack } from '../lib/cdk-app';
import { CertificateStack } from '../lib/cdk-certificate';

const app = new cdk.App();

const stackPrefix = 'fraudmarc-ce-';

/*
 * DomainStack configures the domain's hosting zone and sets up user
 * authentication with Cognito. It establishes the base for the application's
 * security and domain management.
 */
const domainStack = new DomainCdkStack(app, `${stackPrefix}domain`, {
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION
  },
});

/*
 * CertificateStack generates the HTTPS certificate for CloudFront distributions,
 * adhering to CloudFront's ACM certificate requirement in the 'us-east-1' region.
 * It relies on DomainStack for domain info.
 */
const certStack = new CertificateStack(app, `${stackPrefix}certificate`, {
  crossRegionReferences: true,
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: 'us-east-1' 
  },
  fullDomain: domainStack.fullDomain,
  hostedZoneId: domainStack.hostedZoneId,
});
certStack.addDependency(domainStack);


/*
 * AppStack is the application's backbone, incorporating the HTTPS certificate and
 * domain settings from previous stacks. It establishes the hosting, API gateway,
 * and other resources, forming the core AWS infrastructure.
 */
const appStack = new AppCdkStack(app, `${stackPrefix}app`, {
  certificate: certStack.certificate,
  crossRegionReferences: true,
  env: {
    account: process.env.CDK_DEFAULT_ACCOUNT,
    region: process.env.CDK_DEFAULT_REGION
  },
    fullDomain: domainStack.fullDomain,
    httpApiId: domainStack.httpApiId,
  hostedZoneId: domainStack.hostedZoneId,
});
appStack.addDependency(domainStack);
appStack.addDependency(certStack);
