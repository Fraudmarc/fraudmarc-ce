import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { aws_cognito as cognito } from 'aws-cdk-lib';

const SUBDOMAIN = 'fraudmarc-ce';

export class DomainCdkStack extends cdk.Stack {
  public readonly fullDomain: string;
  public readonly hostedZoneId: string;
  public readonly httpApiId: string;

  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    const domainName = this.getDomainName();
    const adminEmail = this.getAdminEmail();
    this.fullDomain = `${SUBDOMAIN}.${domainName}`;

    // Define the Cognito User Pool
    const userPool = new cognito.UserPool(this, 'Fraudmarc CE users', {
      passwordPolicy: {
        minLength: 8,
        requireDigits: false,
        requireLowercase: false,
        requireUppercase: false,
        requireSymbols: false,
      },
      removalPolicy: cdk.RemovalPolicy.DESTROY,
      selfSignUpEnabled: false,
      signInAliases: {
        email: true,
      },
    });

    this.outputCognitoUserCreationCommand(userPool, adminEmail);

    const userPoolClient = new cognito.UserPoolClient(this, 'Fraudmarc CE client', {
      userPool,
      enableTokenRevocation: true,
    });

    const defaultAuthorizer = new cdk.aws_apigatewayv2_authorizers.HttpUserPoolAuthorizer(
      'CognitoAuthorizer', userPool);

    const httpApi = new cdk.aws_apigatewayv2.HttpApi(this, 'Fraudmarc CE api', {
      defaultAuthorizer,
      corsPreflight: {
        allowOrigins: ['*'],
        allowMethods: [cdk.aws_apigatewayv2.CorsHttpMethod.ANY],
        allowHeaders: ['Content-Type', 'X-Amz-Date', 'Authorization', 'X-Api-Key'],
      },
    });

    this.httpApiId = httpApi.apiId;

    new cdk.CfnOutput(this, 'FrontendConfig', {
      description: 'Configuration for the frontend app',
      value: `Set these variables in frontend/app/src/environments/environment.common.ts:
const apiEndpoint = '${httpApi.url}';
const userPoolId = '${userPool.userPoolId}';
const userPoolClientId = '${userPoolClient.userPoolClientId}';
`,
    });

    const hostedZone = new cdk.aws_route53.PublicHostedZone(this, 'Fraudmarc CE zone', {
      zoneName: this.fullDomain,
    });

    this.hostedZoneId = hostedZone.hostedZoneId;

    if (hostedZone.hostedZoneNameServers) {
      const formattedNameServers = cdk.Fn.join("\n", hostedZone.hostedZoneNameServers);
      new cdk.CfnOutput(this, 'DNS setup instructions', {
        description: 'Instructions to configure your Fraudmarc CE subdomain',
        value: `Add this record to ${domainName}'s DNS:
Name: ${this.fullDomain}
Type: NS
Value: ${formattedNameServers}
`,
      });
    }

  }

  getDomainName(): string {
    // Try to get the domain name from an environment variable
    let domainName = process.env.DOMAIN;

    // If not found, try to get it from CDK context variables
    if (!domainName) {
      domainName = this.node.tryGetContext('domain');
    }

    // If still not found, throw an error with instructions
    if (!domainName) {
      throw new Error(
        'Domain name not specified. Ensure the domain allows adding a new ' +
        'NS record for the "fraudmarc-ce" subdomain. Define the domain ' +
        'name through the "DOMAIN" environment variable or a CDK context ' +
        'variable "domain". Example: "-c domain=yourdomain.com" or ' +
        'export DOMAIN=yourdomain.com in your environment.'
      );
    }

    return domainName;
  }

  getAdminEmail(): string {
    // Try to get the admin email from an environment variable
    let adminEmail = process.env.ADMIN_EMAIL;

    // If not found, try to get it from CDK context variables
    if (!adminEmail) {
      adminEmail = this.node.tryGetContext('adminEmail');
    }

    // If still not found, throw an error with instructions
    if (!adminEmail) {
      throw new Error(
        'Admin email not specified. This email is only used for your login to your new app. ' +
        'Define the admin email through the "ADMIN_EMAIL" environment variable or a CDK context ' +
        'variable "adminEmail". Example: "-c adminEmail=youradminemail@example.com" or ' +
        'export ADMIN_EMAIL=youradminemail@example.com in your environment.'
      );
    }

    return adminEmail;
  }

  outputCognitoUserCreationCommand(userPool: cognito.UserPool, adminEmail: string) {
    const cognitoCommand = `
aws cognito-idp admin-create-user \\
--user-pool-id ${userPool.userPoolId} \\
--username ${adminEmail} \\
--user-attributes Name=email,Value=${adminEmail} Name=email_verified,Value=true
`;

    new cdk.CfnOutput(this, 'AdminUserCreationCommand', {
      description: 'Run this command in your terminal to provision your admin user.',
      value: cognitoCommand,
    });
  }

}
