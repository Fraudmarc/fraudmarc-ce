import * as cdk from 'aws-cdk-lib';
import * as acm from 'aws-cdk-lib/aws-certificatemanager';
import { Construct } from 'constructs';

interface CertificateStackProps extends cdk.StackProps {
  fullDomain: string;
  hostedZoneId: string;
}

export class CertificateStack extends cdk.Stack {
  public readonly certificate: acm.Certificate;

  constructor(scope: Construct, id: string, props: CertificateStackProps) {
    super(scope, id, {
      env: { region: 'us-east-1' }, // Ensure this stack is deployed to us-east-1
      ...props,
    });

    const hostedZone = cdk.aws_route53.HostedZone.fromHostedZoneAttributes(this, 'HostedZone', {
      hostedZoneId: props.hostedZoneId,
      zoneName: props.fullDomain,
    });

    this.certificate = new acm.Certificate(this, 'certificate', {
      domainName: props.fullDomain,
      validation: acm.CertificateValidation.fromDns(hostedZone),
    });
  }
}
