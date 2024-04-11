import * as cdk from 'aws-cdk-lib';
import { Construct } from 'constructs';
import { custom_resources as customResources } from 'aws-cdk-lib';
import { NatAsgProvider } from 'cdk-nat-asg-provider';


// import * as sqs from 'aws-cdk-lib/aws-sqs';
import * as lambdaGo from '@aws-cdk/aws-lambda-go-alpha';

// Define constants for database name and user
const DATABASE_NAME = 'dmarc';
const DATABASE_USER = 'postgres';
const BUCKET_PREFIX = 'emails/';
const EMAIL = 'rua';
const TTL = 300; // TTL for the DNS records in seconds

interface GoProps {
  arnLambdaDmarcARResolveBulk?: string;
  bucketName?: string;
  entry: string;
  secretArn: string;
  vpc: cdk.aws_ec2.Vpc;
}

interface AppCdkStackProps extends cdk.StackProps {
  hostedZoneId: string;
  httpApiId: string;
  fullDomain: string;
  certificate: cdk.aws_certificatemanager.Certificate;
}

export class AppCdkStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: AppCdkStackProps) {
    super(scope, id, props);

    const vpc = new cdk.aws_ec2.Vpc(this, 'VPC', {
      maxAzs: 2,
      natGateways: 1,
      natGatewayProvider: new NatAsgProvider({
        instanceType: new cdk.aws_ec2.InstanceType('t3.micro'),
      }),
    });

    // create s3 bucket to receive mail
    const rxBucket = new cdk.aws_s3.Bucket(this, 'rua', {
      autoDeleteObjects: true,
      blockPublicAccess: cdk.aws_s3.BlockPublicAccess.BLOCK_ALL,
      encryption: cdk.aws_s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // create pgsql db
    const databaseInstance = new cdk.aws_rds.DatabaseInstance(this,
      'DbInstance',
      {
        allocatedStorage: 20,
        backupRetention: cdk.Duration.days(1),
        deleteAutomatedBackups: true,
        databaseName: DATABASE_NAME,
        engine: cdk.aws_rds.DatabaseInstanceEngine.postgres({
          version: cdk.aws_rds.PostgresEngineVersion.VER_16,
        }),
        instanceType: new cdk.aws_ec2.InstanceType('t4g.micro'),
        maxAllocatedStorage: 100,
        removalPolicy: cdk.RemovalPolicy.DESTROY,
        storageEncrypted: true,
        storageType: cdk.aws_rds.StorageType.GP2,
        vpc,
        vpcSubnets: {
          subnetType: cdk.aws_ec2.SubnetType.PRIVATE_WITH_EGRESS,
        },
      });

    // Database secret for credentials
    const databaseSecret = databaseInstance.secret!;

    // create lambda to erase and reinitialize db
    const eraseAndReinitDBLambda = this.createGoLambda('eraseAndReinitDB', {
      entry: '../backend/functions/erase_and_reinit_db',
      secretArn: databaseSecret.secretArn,
      vpc,
    });
    databaseInstance.connections.allowDefaultPortFrom(eraseAndReinitDBLambda);
    databaseSecret.grantRead(eraseAndReinitDBLambda);

    const customResourceLambdaRole = new cdk.aws_iam.Role(this,
      'CustomResourceLambdaRole',
      {
        assumedBy: new cdk.aws_iam.ServicePrincipal('lambda.amazonaws.com'),
        description: 'Role for custom resource Lambda to invoke other Lambdas',
      });

    // Granting permissions to invoke the specific Lambda function
    customResourceLambdaRole.addToPolicy(new cdk.aws_iam.PolicyStatement({
      actions: ['lambda:InvokeFunction'],
      resources: [eraseAndReinitDBLambda.functionArn],
    }));

    const eraseAndReinitDBCustomResource = new customResources.AwsCustomResource(this,
      'EraseAndReinitDB',
      {
        onCreate: {
          action: 'invoke',
          parameters: {
            FunctionName: eraseAndReinitDBLambda.functionName,
          },
          physicalResourceId: customResources.PhysicalResourceId.of('EraseAndReinitDB'),
          service: 'Lambda',
        },
        policy: customResources.AwsCustomResourcePolicy.fromSdkCalls({
          resources: customResources.AwsCustomResourcePolicy.ANY_RESOURCE,
        }),
        role: customResourceLambdaRole,
      });

    // Ensure the custom resource is created after the Lambda function is deployed
    eraseAndReinitDBCustomResource.node.addDependency(eraseAndReinitDBLambda);

    // create lambda function to process mail
    const processLambda = this.createGoLambda('process', {
      bucketName: rxBucket.bucketName,
      entry: '../backend/functions/process',
      secretArn: databaseSecret.secretArn,
      vpc,
    });
    databaseInstance.connections.allowDefaultPortFrom(processLambda);
    databaseSecret.grantRead(processLambda);
    rxBucket.grantRead(processLambda);

    // create lambda function to receive mail
    const receiveLambda = this.createGoLambda('receive', {
      arnLambdaDmarcARResolveBulk: processLambda.functionArn,
      bucketName: rxBucket.bucketName,
      entry: '../backend/functions/receive',
      secretArn: databaseSecret.secretArn,
      vpc,
    });
    databaseInstance.connections.allowDefaultPortFrom(receiveLambda);
    databaseSecret.grantRead(receiveLambda);
    processLambda.grantInvoke(receiveLambda);
    rxBucket.grantRead(receiveLambda);

    // create lambda function to serve the API
    const apiServerLambda = this.createGoLambda('api', {
      entry: '../backend/functions/server',
      secretArn: databaseSecret.secretArn,
      vpc,
    });
    databaseInstance.connections.allowDefaultPortFrom(apiServerLambda);
    databaseSecret.grantRead(apiServerLambda);

    // import httpApi from props.httpApi
    const httpApi = cdk.aws_apigatewayv2.HttpApi.fromHttpApiAttributes(this,
      'Fraudmarc-CE API',
      {
        httpApiId: props.httpApiId,
      });

    new cdk.aws_apigatewayv2.HttpRoute(this, 'route', {
      httpApi: httpApi,
      integration: new cdk.aws_apigatewayv2_integrations.HttpLambdaIntegration(
        'apiLambda',
        apiServerLambda,
      ),
      routeKey: cdk.aws_apigatewayv2.HttpRouteKey.with('/{proxy+}',
        cdk.aws_apigatewayv2.HttpMethod.GET),
    });

    const hostedZone = cdk.aws_route53.HostedZone.fromHostedZoneAttributes(this,
      'HostedZone',
      {
        hostedZoneId: props.hostedZoneId,
        zoneName: props.fullDomain,
      });

    new cdk.aws_route53.MxRecord(this, 'SES-MX-Record', {
      ttl: cdk.Duration.seconds(TTL),
      values: [
        {
          priority: 10,
          hostName: `inbound-smtp.${this.region}.amazonaws.com`,
        },
      ],
      zone: hostedZone,
    });

    new cdk.aws_route53.TxtRecord(this, 'DMARCAuthorizationForAll', {
      recordName: `*._report._dmarc.${props.fullDomain}`,
      ttl: cdk.Duration.seconds(TTL),
      values: ['v=DMARC1'],
      zone: hostedZone,
    });

    // link hosted zone to SES
    const sesDomain = new cdk.aws_ses.EmailIdentity(this, 'Fraudmarc CE domain', {
      identity: cdk.aws_ses.Identity.publicHostedZone(hostedZone),
    });

    // set ses to receive mail
    const ruleSet = new cdk.aws_ses.ReceiptRuleSet(this, 'Fraudmarc CE RuleSet');

    // Construct the ReceiptRuleSet ARN
    const receiptRuleSetArn = this.formatArn({
      account: this.account,
      region: '', // SES ARN does not require a region; leave this empty
      resource: `receipt-rule-set/${ruleSet.receiptRuleSetName}`,
      service: 'ses',
    });

    new cdk.aws_ses.ReceiptRule(this, 'Fraudmarc-CE-Rule', {
      actions: [
        // Action to store the email in S3
        new cdk.aws_ses_actions.S3({
          bucket: rxBucket,
          objectKeyPrefix: BUCKET_PREFIX,
        }),
        // Action to trigger the Lambda function
        new cdk.aws_ses_actions.Lambda({
          function: receiveLambda,
          invocationType: cdk.aws_ses_actions.LambdaInvocationType.EVENT,
        }),
      ],
      enabled: true,
      recipients: [`${EMAIL}@${props.fullDomain}`],
      ruleSet,
    });

    // Grant SES permission to invoke the Lambda function
    receiveLambda.addPermission('AllowSESInvocation', {
      principal: new cdk.aws_iam.ServicePrincipal('ses.amazonaws.com'),
      sourceAccount: this.account,
    });

    // Create a custom resource to activate an deactivate the rule set
    const ruleSetActivation = new customResources.AwsCustomResource(this,
      'RuleSetActivation',
      {
        onCreate: {
          service: 'SES',
          action: 'setActiveReceiptRuleSet',
          parameters: {
            RuleSetName: ruleSet.receiptRuleSetName,
          },
          physicalResourceId: customResources.PhysicalResourceId.of('RuleSetActivation'),
        },
        onDelete: {
          service: 'SES',
          action: 'setActiveReceiptRuleSet',
          parameters: {
            // Intentionally left blank to deactivate the rule set
          },
          physicalResourceId: customResources.PhysicalResourceId.of('RuleSetActivation'),
        },
        policy: customResources.AwsCustomResourcePolicy.fromSdkCalls({
          resources: customResources.AwsCustomResourcePolicy.ANY_RESOURCE,
        }),
      });

    // Ensure the rule set is deactivated before attempting to delete
    ruleSetActivation.node.addDependency(ruleSet);

    // Create a new S3 bucket with public access blocked for the frontend
    const frontendBucket = new cdk.aws_s3.Bucket(this, 'FrontendBucket', {
      autoDeleteObjects: true,
      blockPublicAccess: cdk.aws_s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Create a CloudFront distribution for the frontend S3 bucket
    const frontendDistribution = new cdk.aws_cloudfront.Distribution(this,
      'FrontendDistribution',
      {
        certificate: props.certificate,
        defaultBehavior: {
          origin: new cdk.aws_cloudfront_origins.S3Origin(frontendBucket),
          viewerProtocolPolicy: cdk.aws_cloudfront.ViewerProtocolPolicy.HTTPS_ONLY,
        },
        defaultRootObject: 'index.html',
        domainNames: [props.fullDomain],
        errorResponses: [
          {
            httpStatus: 403, // Access Denied
            responseHttpStatus: 200,
            responsePagePath: '/index.html',
            ttl: cdk.Duration.seconds(300),
          },
          {
            httpStatus: 404, // Not Found
            responseHttpStatus: 200,
            responsePagePath: '/index.html',
            ttl: cdk.Duration.seconds(300),
          },
        ],
      });

    // Create DNS records for the CloudFront distribution using Route53
    new cdk.aws_route53.ARecord(this, 'FrontendAliasRecord', {
      target: cdk.aws_route53.RecordTarget.fromAlias(
        new cdk.aws_route53_targets.CloudFrontTarget(frontendDistribution)
      ),
      zone: hostedZone,
    });

    // Deploy the frontend application to the S3 bucket
    new cdk.aws_s3_deployment.BucketDeployment(this, 'DeployFrontendApp', {
      destinationBucket: frontendBucket,
      distribution: frontendDistribution,
      distributionPaths: ['/*'],
      sources: [cdk.aws_s3_deployment.Source.asset('../frontend/dist')],
    });

    // Output the usage instructions for the application
    this.outputAppUsage(props.fullDomain);

  }

  createGoLambda(name: string, props: GoProps): lambdaGo.GoFunction {
    return new lambdaGo.GoFunction(this, name, {
      architecture: cdk.aws_lambda.Architecture.ARM_64,
      bundling: {
        environment: {
          GOARCH: 'arm64',
        },
      },
      entry: props.entry,
      environment: {
        ArnLambdaDmarcARResolveBulk: props.arnLambdaDmarcARResolveBulk ?? '',
        ARRTable: 'AggregateReportRecord',
        ARTable: 'AggregateReport',
        BUCKET_NAME: props.bucketName ?? '',
        BUCKET_PREFIX: BUCKET_PREFIX,
        DRE_TABLE: 'dmarc_reporting_entries',
        REPORTING_DB_HOST: `{{resolve:secretsmanager:${props.secretArn}:SecretString:host}}`,
        REPORTING_DB_NAME: DATABASE_NAME,
        REPORTING_DB_USER: DATABASE_USER,
        REPORTING_DB_PASSWORD: `{{resolve:secretsmanager:${props.secretArn}:SecretString:password}}`,
        REPORTING_DB_MAX_TIME: '1h',
        REPORTING_DB_SSL: 'require',
      },
      memorySize: 1024,
      runtime: cdk.aws_lambda.Runtime.PROVIDED_AL2023,
      retryAttempts: 0,
      timeout: cdk.Duration.seconds(300),
      vpc: props.vpc,
      vpcSubnets: {
        subnetType: cdk.aws_ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
    });
  };

  outputAppUsage(fullDomain: string) {
    const usage = `DMARC rua address: ${EMAIL}@${fullDomain}
Fraudmarc CE url: https://${fullDomain}/
`;

    new cdk.CfnOutput(this, 'AppUsage', {
      description: 'Instructions for using the Fraudmarc CE application.',
      value: usage,
    });
  }

}
