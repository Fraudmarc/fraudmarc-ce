<img src="25Fraudmarc-CE-Logo-on-Light.png" alt="logo">

# Installation Guide for Fraudmarc Community Edition :fire: :fire: :fire:

Here are the steps to setup Fraudmarc CE to collect and process DMARC data for your domain(s). 

1. Create AWS role for Fraudmarc CE
2. Create a PostgreSQL Database
3. Setup Docker
4. Deploy Lambda functions to process DMARC reports with Docker Hub
5. Configure AWS SES to receive DMARC reports & invoke our Lambdas for processing
6. Run the Fraudmarc CE client with Docker
7. Publish a DMARC policy to collect reports

## System overview

![IDiagram](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/diagram2.png)


## Let's get started

**Want DMARC data without complex cloud infrastructure? Try our [hosted DMARC service](https://www.fraudmarc.com/plans/).**

### 1. Create AWS Group and User for Fraudmarc CE👍
**This installation guide is aimed for users who want to process DMARC reports quickly and easily. If you want to setup SES, S3 bucket, RDS instance, Lambda functions yourself, follow [this guide](). Otherwise, please understand that in order for you to setup the Fraudmarc CE with the minimum amount of effort, you need to grant us with certain permissions in order to setup the different AWS services for you.**
*Before proceeding the install guide, create an AWS account!*
We will first start with creating an AWS Group and adding a User to the group that will be used to setup the AWS services automatically.

1. Go to the IAM service in the AWS Console. In the left nav bar, click `Groups` and `Create New Group`. You can name the group however you like. Proceed to `Next Step`.

2. Using the filter, search and click the checkbox next to the following policies:
* AmazonRDSFullAccess (grants access to create DB instance for you)
* AWSLambdaFullAccess (grants access to create Lambda Functions for you)
* IAMFullAccess (grants access to create Roles, and add policies for you)
* AmazonS3FullAccess (grants access to create S3 bucket for you)
* AmazonSESFullAccess (grants access to setup SES that will receive DMARC reports)
* CloudWatchLogsFullAccess (not used in setup, but you can use it to find problems/bugs)

*you can check our Dockerfiles in our `root` directory and in `/installer/` to see what we actually do with these privileges. We are only using these to help you setup your Fraudmarc CE, and if in doubt, you can perform the installation steps yourself*

3. Review and `Create Group`

4. Go to the `Users` tab and `Add user`. You can name the user however you like (i.e. fraudmarc-ce-installer). Click on the `programmatic access` in `Access type` and click `Next:Permissions`.

5. Inside the `Add user to group` box, check the box next to the new group you created and proceed to `Next:Review`. After reviewing, click `Create user`.

6. You will now see values inside `Access key ID` and `Secret access key`. Copy those values into the `/installer/env.list` file as such:

```
...
AWS_ACCESS_KEY_ID=YOURACCESSKEY
AWS_SECRET_ACCESS_KEY=YourSecretKey
...
```
7. Choose a name for your S3 Bucket(needs to be a globally unique name) that will receive DMARC report emails from SES, and enter the name into `/installer/env.list` as such:

```
...
BUCKET_NAME=My-Globally-Unique-Bucket-Name
...
```
This is needed for us to deploy your lambda functions.

8. Hooray! The hard part is done :)

### 1. Create AWS Role for Fraudmarc CE👍

1. Go to AWS IAM panel, and click the Roles on the left. Create a role for Fraudmarc CE lambda functions. 

2. Select AWS Service, and click lambda function.

3. Check the box net to the `AmazonS3ReadOnlyAccess` and  `CloudWatchLogsFullAccess` permissions.

4. Set the Role name to `FraudmarcCE` and Create role.

5. Click on the role you just created to see the Role ARN in the form like `arn:aws:iam::<AccountID>:role/fraudmarcce`. Copy this ARN to the existing role field in the installer/env.list (don't be confused with the env.list in the root directory!) ⬇️

   ```
   AWS_ROLE_ARN=`arn:aws:iam::<AccountID>:role/fraudmarcce
   ```

### 2. Setup Your Database:thumbsup:

Instructions for creating the database in AWS RDS. You are welcome to use any other PostgreSQL database server.

*If you need maximum security, use a private VPC for your RDS instance and Lambdas. Such configuration is beyond the scope of this document.* 

1. Set up a RDS PostgreSQL database via [AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_GettingStarted.CreatingConnecting.PostgreSQL.html) instruction

   1. Go to AWS RDS panel and click the region on the upper right corner, and choose one of the following based on your position: `US East N.V`, `EU (Ireland)`, `US West (OR)`.
   2. Go to AWS RDS panel and click Instances on the left panel, and click Launch DB instance
   3. Through the creating process, you need to choose PostgreSQL, decide the DB instance class based on your DMARC report volume, and fill the DB instance identifier, master username and password.
   4. Set Public accessibility to Yes.
   5. Set the name of Database to `fraudmarcce`

2. After you launch the DB instance, you need to wait a few minutes for it to completely setup (check the `DB instance status` in the Summary panel). 

3. If you have already configured your AWS account jump to `Step 3`. Otherwise, find your instance's `Availability zone` (i.e.  `us-east-1`). Create a new file named `config` under `~/.aws`, and the content should be as such:

   ```
   [default]
   region=[your region]
   ```

4. Choose the instance you just created, scroll down to Security group rule, and click the `Inbound` tab at the bottom of the new window.

5. Click the `Edit` button, and click the `Add rule` button. Give the Port Range the same with the previous one, and Source should be `Anywhere` from the drop down.

6. Install the [PSQL](<https://www.postgresql.org/download/>) command line tool on your local machine.

7. Import the Fraudmarc CE schema into your new database with the command (You can find your endpoint on your RDS instance panel) below: 

   ```shell
   cd /path/to/fraudmarc-ce
   pg_restore --no-privileges --no-owner -v -h [endpoint of instance] -U [master username] -n public -d [new database name (not instance name)] fraudmarcce
   ```

8. Go to RDS panel and choose the Instances on the left panel. Choose the `fraudmarcce` instance you just created. Copy the `DB name`, `Username`, `Endpoint` to the `env.list` file in the project repository root directory like:arrow_down:

   ```
   # Set to match your environment
   REPORTING_DB_NAME=[DB name]
   REPORTING_DB_USER=[Username]
   REPORTING_DB_PASSWORD=[password]
   REPORTING_DB_HOST=[Endpoint]
   REPORTING_DB_SSL=require
   REPORTING_DB_MAX_TIME=180s
   ```

### 3. Setup Docker and Pull Install Docker Image
This step runs the fraudmarc-ce-install image, which using the given Role ARN and AWS credentials, deploys the Lambda functions that are needed to retrieve and process the DMARC reports from the S3 bucket and update the Postgres Database.
1. Run the command to update docker:

   ```shell
   sudo apt install docker.io
   ```
2. Run Docker Image from Docker Hub
*If the following commands prompts permission denied error on your computer, you may need `sudo` privilege.*

You are able to run the install Docker image from the Docker Hub public repository, or you can build it from scratch using our Dockerfile in the `installer` directory.

* run Docker image from public repo:
    ```shell
    docker run -it --env-file env.list --env-file installer/env.list fraudmarc/fraudmarc-ce-install
    ```
* build and run Docker image from /installer/Dockerfile:
    ```shell
    docker build -t fraudmarc-ce-install -f installer/Dockerfile .
    docker run -it --env-file env.list --env-file installer/env.list fraudmarc-ce-install
    ```

### 4. Setup Your AWS Simple Email Service (SES):thumbsup:

1. In AWS SES, choose the Domain on the left side, and click Verify a New Domain. Enter  `fraudmarc-ce.<your domain name>`. Your DMARC Reports will be delievered here. Follow the instructions to setup the Domain Verification Record and Email Receiving Record to complete the verification.

   > If you use the service provider like Godaddy, add a new DNS TXT record with the host field `_amazonses.fraudmarc-ce` (without your domain name) and the TXT field from AWS. Add a new MX record with `fraudmarc-ce` as the host field and the MX field from AWS without the '10'. If your domain service provider requires you to configure the mx record priority, place the '10' in the priority field.

2. Choose the Rule Sets on the left panel. If you are new to AWS, create a Receipt Rule. If you have existing rules, create a new rule.

3. Enter the email address that the DMARC Report will be sent to (`dmarc@fraudmarc-ce.<your domain name>`), and click Next Step.

4. Add action with type S3, and create a S3 bucket with a globally unique name.

5. Add action with type lambda and chose your receive lambda function. If prompted, grant permissions.

6. Copy the S3 bucket name to the `project.json` file in project repository like:arrow_down:

   ```json
   ...
   "environment": {
       "BUCKET_NAME": "[bucket name]",
   },
   ```

### Run The Fraudmarc CE Docker:thumbsup:

We've simplified the client side of Fraudmarc CE by providing a single Docker to provide both the Angular frontend  and Go backend.

3. Navigate to the directory containing the Fraudmarc CE docker image.

4. Run the command to download and set up dependencies. This process may take several minutes

   ```shell
   docker build -t fraudmarc-ce .
   ```

5. Run the docker image

   ```shell
   docker run -it --env-file env.list -p 7489:7489 fraudmarc-ce
   ```

6. Your Fraudmarc CE installation is now ready at [http://localhost:7489](http://localhost:7489).

### Create a DMARC policy:thumbsup:

**It may take between a few hours and a few days for your first data to arrive.** 

Be sure you have created a DMARC policy in your domain's DNS so that DMARC aggregate reports will be sent to the address you set in SES step 3 above.

Your DMARC record should be similar to:

   - **Type**: `TXT`
   - **Hostname**: `_dmarc`
   - **Value**: `v=DMARC1; p=none; rua=mailto:dmarc@fraudmarc-ce.<your domain name>; ri=3600;`
       
### :grimacing: Questions?

If you have questions about installation or email authentication, please reference the Fraudmarc Community forum at [https://community.fraudmarc.com](https://community.fraudmarc.com).
