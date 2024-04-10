
# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.1.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [2.0.0] - 2024-04-10
### Added
- Multi-User Support through AWS Cognito for secure and easy collaboration.
- Custom Domain Hosting integration with CloudFront, S3, and ACM certificates for a personalized experience.
- Serverless Backend leveraging AWS API Gateway and Lambda for enhanced flexibility and scalability.
- Private Database Hosting on RDS free-tier arm64 instance for cost-effective storage.
- Infrastructure Automation via AWS CDK and provisioned through AWS CloudFormation for improved reliability.
- Managed DNS for simplified domain management.
- A single RUA DMARC reporting address that aggregates data across unlimited domains.

### Changed
- Upgraded frontend from Angular v6 to v16, modernizing the stack.
- Transitioned backend from Go 1.10 to Go 1.22, adopting Go modules for better dependency management.
- Enhanced security with Secrets Manager, IAM policies, Security Groups, and private VPC subnets.

## [1.0.0] - 2022-07-06
### Changed
- t3 became new db instance family default

## [1.0.0] - 2018-08-03
### Added
- Commercial feature list

## [0.0.4] - 2018-07-30
### Added
- CHANGELOG containing information on releases by @AbigailCliche

## [0.0.3] - 2018-07-12
### Changed
- Angular package version updates

## [0.0.2] - 2018-06-26
### Changed
- Faster and smaller Docker container build process
- Completed INSTALL and INSTALL-ADVANCED, installation instructions and advanced installation instructions for installation with and without a Docker image by @kimkb2011

## 0.0.1 - 2018-06-14
### Added
- README containing overview of purpose and architecture
- INSTALL and INSTALL-ADVANCED containing setup instructions
- DMARC report data analytics panel
- Domains overview panel
- Ability to view report data between specific date ranges
