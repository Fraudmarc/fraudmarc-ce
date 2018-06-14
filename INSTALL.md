<img src="25Fraudmarc-CE-Logo-on-Light.png" alt="logo">

# Installation Guide for Fraudmarc Community Edition :fire: :fire: :fire:

Here are the steps to setup Fraudmarc CE to collect and process DMARC data for your domain(s). 

1. Create a PostgreSQL Database
2. Create an AWS role for Fraudmarc CE
3. Install Go
4. Install Git
5. Build & deploy AWS Lambda functions to process DMARC reports
6. Configure AWS SES to receive DMARC reports & invoke our Lambda for processing
7. Build & run the Fraudmace CE client
8. Publish a DMARC policy to collect reports

## System overview

![IDiagram](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/diagram2.png)


## Let's get started

**Want DMARC data without complex cloud infrastructure? Try our [hosted DMARC service](https://www.fraudmarc.com/plans/).**

### Set Up Your Database:thumbsup:

Instructions are for creating the database in AWS RDS. You are welcome to use any other PostgreSQL database server.

*If you need maximum security, use a private VPC for your RDS instance and Lambdas. Such configuration is beyond the scope of this document.* 

1. Set up a RDS PostgreSQL database via [AWS](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_GettingStarted.CreatingConnecting.PostgreSQL.html) instruction

   1. Go to AWS RDS panel and click the region on the upper right corner, and choose one of the following based on your position: `US East N.V`, `EU (Ireland)`, `US West (OR)`.
   2. Go to AWS RDS panel and click Instances on the left panel, and click Launch DB instance
   3. Through the creating process, you need to choose PostgreSQL, decide the DB instance class based on your DMARC report volume, and fill the DB instance identifier, master username and password.
   4. Set Public accessibility to Yes.
   5. Set the name of Database to `fraudmarcce`

2. After you launch the DB instance, you need to wait a few minutes for it to completely setup (check the `DB instance status` in the Summary panel). 

3. If you have already configured your AWS account jump to `Step 4`. Otherwise, find your instance's `Availability zone` (i.e.  `us-east-1`). Create a new file named `config` under `~/.aws`, and the content should be as such:

   ```
   [default]
   region=[your region]
   ```

4. Choose the instance you just created, scroll down to Security group rule, and click the `Inbound` tab at the bottom of the new window.

5. Click the `Edit` button, and click the `Add rule` button. Give the Port Range the same with the previous one, and Source should be `Anywhere`

6. Install the [PSQL](<https://www.postgresql.org/download/>) command line tool on your local machine.

7. Import the Fraudmarc CE schema into your new database with the command (You can find your endpoint on your RDS instance panel) below: 

   ```shell
   cd /path/to/fraudmarc-ce
   pg_restore --no-privileges --no-owner -v -h [endpoint of instance] -U [master username] -n public -d [new database name (not instance name)] fraudmarcce
   ```

8. Go to RDS panel and choose the Instances on the left panel. Choose the `fraudmarcce` instance you just created. Copy the `DB name`, `Username`, `Endpoint` to the `project.json` file in project repository like:arrow_down:

   ```json
   ...
   "environment": {
       "REPORTING_DB_NAME": "[DB name]",
       "REPORTING_DB_USER": "[Username]",
       "REPORTING_DB_PASSWORD": "[password]",
       "REPORTING_DB_HOST": "[Endpoint]",
       "REPORTING_DB_SSL": "require",
       "REPORTING_DB_MAX_TIME": "180s",
   },
   ...
   ```

9. Copy the information in Step 5 to the env.list in the project repository like:arrow_down:

   ```
   # Set to match your environment
   REPORTING_DB_NAME=[DB name]
   REPORTING_DB_USER=[Username]
   REPORTING_DB_PASSWORD=[password]
   REPORTING_DB_HOST=[Endpoint]
   REPORTING_DB_SSL=require
   REPORTING_DB_MAX_TIME=180s
   ```

### Creat AWS Role for Fraudmarc CE👍

1. Go to AWS IAM panel, and click the Roles on the left. Create a role for fraudmarc CE lambda functions.

2. Select AWS Service, and click lambda function.

3. Check the box net to the `AmazonS3ReadOnlyAccess` and  `CloudWatchLogsFullAccess` permissions.

4. Set the Role name to `FraudmarcCE` and Create role.

5. Click on the role you just created to see the Role ARN in the form like `arn:aws:iam::<AccountID>:role/fraudmarcce`. Copy this ARN to the existing role field in project.json in the project repository like ⬇️

   ```json
   {
       "role": "arn:aws:iam::[AccountID]:role/FraudmarcCE"
   }
   ```

### Install GOLANG:thumbsup:

Follow the [Go Installation Steps](https://golang.org/doc/install) to install Go on your machine, and set the `GOPATH` via this [link](https://github.com/golang/go/wiki/SettingGOPATH)

> You may want to export the path to `.bashrc` file

### Install Git:thumbsup:

1. Follow the link to install [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) , and configure the Git by checking this [link](https://help.github.com/articles/setting-your-username-in-git/).

2. After installing the Git, you can use commands:arrow_down: to clone the fraudmarc CE project to your local machine:

   ```shell
   go get github.com/fraudmarc/fraudmarc-ce
   ```



### Deploy Your Lambda Function:thumbsup:

1. If you use macOS, Linux, or OpenBSD, run the command:arrow_down: to install CURL on your computer:

   ```shell
   sudo apt install curl
   ```

2. Follow the [APEX](http://apex.run/) instructions to install apex.

3. Follow the [AWS Credentials](https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html) to generate access key and ID.

   1. If you have never configured AWS credentials before, create an `~/.aws` directory and a `credentials` file under that directory containing:

   ```
   [default]
   aws_access_key_id=[access key]
   aws_secret_access_key=[secret key]
   ```

   1. If you have previously created credentials, add the new ones to your existing file.

4. Run the following commands to install dependencies:arrow_down:

   ```shell
   go get github.com/fraudmarc/fraudmarc-ce/backend/lib \
          github.com/fraudmarc/fraudmarc-ce/database
   (go get -d gopkg.in/mgutz/dat.v1 ; exit 0)
   cd $GOPATH/src/gopkg.in/mgutz/dat.v1
   patch -p1 < $GOPATH/src/github.com/fraudmarc/fraudmarc-ce/database/dat.patch
   cd $GOPATH/src/github.com/fraudmarc/fraudmarc-ce
   go get ./...
   ```   

5. Build and deploy the two lambda functions

    ```shell
    apex deploy
    ```

6. Go to AWS Console lambda function and copy the process function's ARN. Open `project.json` in the code, change

   ```json
   {
   	"ArnLambdaDmarcARResolveBulk": "***YOUR PROCESS LAMBDA FUNCTION ARN***"
   }
   ```

7. Go to AWS IAM panel, click on the role you just created and choose Add inline policy at the lower right corner. Click the JSON, replace the content with⬇️

   ```json
   {
       "Version": "2012-10-17",
       "Statement": [
           {
               "Sid": "VisualEditor0",
               "Effect": "Allow",
               "Action": [
                   "lambda:InvokeFunction",
                   "lambda:InvokeAsync"
               ],
               "Resource": "***YOUR PROCESS LAMBDA FUNCTION ARN***"
           }
       ]
   }
   ```

8. Re-deploy the lambdas with the updated configuration that to point to the correct endpoint.

    ```shell
    apex deploy
    ```

### Set Up Your AWS Simple Email Service (SES):thumbsup:

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

1. Run the command to update docker:

   ```shell
   sudo apt install docker.io
   ```

2. If the following commands prompts permission denied error on your computer, you may need `sudo` privilege.

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
