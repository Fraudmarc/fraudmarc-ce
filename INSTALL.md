# Installation Guide for Fraudmarc CE

Welcome to the installation guide for Fraudmarc Community Edition. This guide will walk you through the process of deploying Fraudmarc CE to your AWS account, configuring your domain, and setting up the necessary components to start monitoring DMARC aggregate reports.

## Prerequisites

Before you begin, ensure you have the following:

1. **Golang**: Install from [go.dev](https://go.dev), version 1.22.1 confirmed working.
2. **Node.js**: Install version 18 (LTS) from [Node.js](https://nodejs.org/), version 18.20.1 confirmed working.
3. **AWS Account**: With permissions to deploy CDK stacks.
4. **AWS CLI**: Installed and configured as per [AWS CLI installation](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).

## Installation Steps

### Step 1: Deploy Domain Stack

Navigate to the `cdk` directory and execute the following commands. This will install build dependencies then deploy resources necessary for your domain and authentication. This process may take a few minutes.

```sh
npx yarn
npx aws-cdk -c domain=example.com -c adminEmail=admin@example.com --require-approval never deploy fraudmarc-ce-domain
```

*Replace `example.com` and `admin@example.com` with your actual domain and admin email address.*

#### Expected Output

You will receive output similar to the following, which includes important information such as the admin user creation command, DNS setup instructions, and frontend configuration details:

```
✅  fraudmarc-ce-domain

Outputs:
fraudmarc-ce-domain.AdminUserCreationCommand = 
aws cognito-idp admin-create-user \
--user-pool-id us-east-1_XXXXXXXXX \
--username admin@example.com \
--user-attributes Name=email,Value=admin@example.com Name=email_verified,Value=true

fraudmarc-ce-domain.DNSsetupinstructions = Add this record to example.com's DNS:
Name: fraudmarc-ce.example.com
Type: NS
Value: ns-XXXX.awsdns-XX.org, ns-XXXX.awsdns-XX.com, etc.

fraudmarc-ce-domain.FrontendConfig = Set these variables in frontend/app/src/environments/environment.common.ts:
const apiEndpoint = 'https://XXXXXXXXXX.execute-api.us-east-1.amazonaws.com/';
const userPoolId = 'us-east-1_XXXXXXXXX';
const userPoolClientId = 'XXXXXXXXXXXXXXXXXXXXXXXXXX';

Stack ARN: arn:aws:cloudformation:us-east-1:XXXXXXXXXXXX:stack/fraudmarc-ce-domain/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
```

### Step 2: Configure DNS

Add at least one of the NS records provided in the output of Step 1 to your domain registrar's DNS management page. Depending on your registrar, you may be able to add all nameservers into a single resource record, or you might need to create separate records for each nameserver. It's essential to install at least one to ensure proper DNS resolution for your domain.

### Step 3: Create Admin User

Execute the admin user creation command from Step 1's output. You will receive an email with a temporary password for your first login.

### Step 4: Configure Frontend

Open `frontend/app/src/environments/environment.common.ts` in a text editor and set the variables as provided in the `FrontendConfig` output from Step 1. This connects your frontend app to the backend services.

### Step 5: Build Frontend

While DNS changes are propagating, proceed to build the frontend application:

```sh
cd ../frontend
npx yarn
npx @angular/cli@16 build --configuration=production
```

### Step 6: Deploy Main App Stack

Return to the `cdk` directory and deploy the main app stack. This includes setting up a CloudFront distribution, which requires the certificate stack to be installed in `us-east-1`. This process may take up to 15 minutes.

```sh
npx aws-cdk -c domain=example.com -c adminEmail=admin@example.com --require-approval never deploy fraudmarc-ce-app
```

*Monitor the deployment closely, especially during the certificate creation phase. If it appears stuck, verify that your NS records are correctly set as per Step 2.*

#### Expected Output

Upon successful deployment, you will receive output similar to:

```
✅  fraudmarc-ce-app

Outputs:
fraudmarc-ce-app.AppUsage = DMARC rua address: rua@fraudmarc-ce.example.com
Fraudmarc CE url: https://fraudmarc-ce.example.com/

Stack ARN: arn:aws:cloudformation:us-east-1:XXXXXXXXXXXX:stack/fraudmarc-ce-app/XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
```

### Step 7: Configure DMARC Reporting

Add the DMARC rua email address from Step 6's output (e.g., `rua@fraudmarc-ce.example.com`) to the DMARC record for each domain you want to analyze. Consider setting the reporting interval (`ri`) to a low value (e.g., `3600` seconds) to receive reports more frequently. For example:

```
v=DMARC1; p=none; rua=mailto:rua@fraudmarc-ce.example.com; ri=3600;
```

## Conclusion

Congratulations! You have successfully installed Fraudmarc Community Edition. It may take some time for mailbox providers to start sending DMARC aggregate reports to your new address. Monitor the system and verify its operation as reports begin to arrive.

For support and further assistance, visit the [Fraudmarc CE GitHub repository](https://github.com/Fraudmarc/Fraudmarc-CE).