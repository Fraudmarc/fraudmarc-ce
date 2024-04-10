![Fraudmarc CE Logo](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/Fraudmarc-CE-Logo-on-Light.png)

# Fraudmarc Community Edition (CE)

- **Empowering Federal Agencies & Beyond:** Trusted by government bodies for its unwavering commitment to data control.
- **Years of Excellence:** Proven reliability, serving agencies, businesses, and individuals since 2018.
- **Data Localization Compliance:** Easily satisfies data residency requirements by specifying the desired AWS region.
- **Fully Open Source:** Complete transparency and control.
- **Effortlessly Scalable:** Built on AWS, ensuring seamless scalability and performance.
- **Simplicity Meets Efficacy:** Easy setup for immediate impact.

## Elevate Your DMARC Insight

Fraudmarc CE v2 offers a secure, scalable system to analyze DMARC aggregate reports, unlocking valuable insights into all email activities associated with your domain. Designed with government agencies in mind, it meets stringent data control policies without sacrificing ease of use or scalability.

### Why Fraudmarc CE?

- **Centralized DMARC Reporting:** One rua DMARC reporting address collects data across unlimited domains, simplifying management and visibility.
- **Enhanced User Experience:** With multi-user support and authentication through AWS Cognito, teams can collaborate efficiently.
- **Robust Infrastructure:** Leveraging AWS services like CloudFront, S3, API Gateway, Lambda, and RDS, coupled with IAM policies and private VPC subnets, ensures unmatched security and reliability.
- **Modernized Technologies:** Transition to Angular v16 and Go 1.22 facilitates a more responsive frontend and an efficient, modular backend.

## Getting Started

Dive into DMARC data with ease:

1. **Setup Made Simple:** Our serverless architecture on AWS and AWS GovCloud means you focus on insights, not infrastructure.
2. **Secure & Scalable:** From receiving DMARC reports with AWS SES to processing them with Lambda functions, your data and infrastructure is secure and scalable.

See our [Installation Guide](INSTALL.md) for step-by-step installation instructions.

## What's New in v2?

- **Multi-User Support:** Seamless collaboration with AWS Cognito authentication.
- **Custom Domain Hosting:** Use your domain with CloudFront, S3, and ACM certificates for a branded experience.
- **Serverless Backend:** Powered by AWS API Gateway and Lambda for flexibility and scale.
- **Private Database Hosting:** On RDS free-tier arm64 instance for cost-effective storage.
- **Enhanced Security:** Using Secrets Manager, IAM policies, Security Groups, & private VPC subnets.
- **Infrastructure Automation:** Defined by AWS CDK and provisioned through AWS CloudFormation for reliability.
- **Managed DNS:** Simplifies domain management.
- **Modernized Stack:** From Angular v6 to v16 and Go 1.10 to Go 1.22 for frontend and backend upgrades.

## Unlock Your DMARC Data

Gain insights into all email sources from your domain, enhancing your security and authentication management.

![Introduction](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/newgif.gif)

## Architectural Overview

![Architecture Diagram](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/diagram2.png)

Built on a robust AWS foundation, Fraudmarc CE simplifies DMARC report processing while ensuring data security and system scalability. Join our journey to secure email communication across domains, embracing the future of DMARC data analysis with Fraudmarc CE v2.

*Explore our [DMARC services](https://www.fraudmarc.com/plans/) for an even easier way to manage email authentication across your business.*
