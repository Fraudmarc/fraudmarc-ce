<img src="25Fraudmarc-CE-Logo-on-Light.png" alt="logo">

# Installation Guide for Fraudmarc Community Edition :fire: :fire: :fire:

Here are the steps to setup Fraudmarc CE to collect and process DMARC data for your domain(s). 

1. Create AWS User for Fraudmarc CE
2. Run the Fraudmarc CE Install Docker
3. Configure Your PostgreSQL Database
4. Setup Your AWS Simple Email Service (SES)
5. Run The Fraudmarc CE Client Docker
6. Publish a DMARC policy to collect reports

## System overview

![IDiagram](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/diagram2.png)


## Let's get started

**Want DMARC data without complex cloud infrastructure? Try our [hosted DMARC service](https://www.fraudmarc.com/plans/).**

### 1. Create AWS Group and User for Fraudmarc CE👍
**This installation guide is aimed for users who want to process DMARC reports quickly and easily. If you want to setup SES, RDS DB(Postgres DB), Lambda functions yourself, follow [this guide](https://github.com/Fraudmarc/fraudmarc-ce/blob/master/ADVANCED_INSTALL.md). Otherwise, please understand that in order for you to setup Fraudmarc CE with the minimum amount of effort, you need to grant us with certain permissions in order to setup the different AWS services for you.**

*Before proceeding the install guide, [create an AWS account](https://aws.amazon.com/free/)!*

*Also, don't forget to clone this repository to your local machine :)*

We will first start with creating an AWS Group and adding a User to the group that will be used to setup the AWS services automatically.

1. Go to the IAM service in the AWS Console. In the left nav bar, click `Groups` and `Create New Group`. You can name the group however you like. Proceed to `Next Step`.

2. Using the filter, search and click the checkbox next to the following policies:
* AmazonRDSFullAccess (grants access to create DB instance for you)
* AWSLambdaFullAccess (grants access to create Lambda Functions for you)
* IAMFullAccess (grants access to create Roles, and add policies for you)
* AmazonS3FullAccess (grants access to create S3 bucket for you)
* AmazonSESFullAccess (grants access to setup SES that will receive DMARC reports)
* CloudWatchLogsFullAccess (not used in setup, but you can use it to find problems/bugs)

*you can check our Dockerfiles in the `root` directory and in `/installer/` to see what we actually do with these privileges. We are only using these to help you setup your Fraudmarc CE, and if in doubt, you can perform the installation steps yourself :) *

3. Review and `Create Group`

4. Go to the `Users` tab and `Add user`. You can name the user however you like (i.e. fraudmarc-ce-installer). Click on the box next to `programmatic access` in `Access type` and click `Next:Permissions`.

5. Inside the `Add user to group` tab, check the box next to the new group you created and proceed to `Next:Review`. After reviewing, click `Create user`.

6. You will now see values inside `Access key ID` and `Secret access key`. Copy those values into `/installer/env.list`:

```
...
AWS_ACCESS_KEY_ID=YOURACCESSKEY
AWS_SECRET_ACCESS_KEY=YourSecretKey
...
```
7. Check out [AWS Regions](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html) and identify your AWS region. This will be your default location. Add the default AWS region (i.e.  `us-east-1`) to `/installer/env.list`:
```
...
AWS_DEFAULT_REGION=your-aws-region
...
```

8. Enter into `/installer/env.list` your S3 Bucket name (needs to be a globally unique name - i.e. fraudmarc-ce-YourCompany) that will receive DMARC report emails from SES:

```
...
BUCKET_NAME=Your-Globally-Unique-Bucket-Name
...
```

### 2. Setup Docker and Run fraudmarc-ce-install Docker Image
This step runs the fraudmarc-ce-install Docker image, which creates an AWS Role, launches a free-tier RDS Postgres database, and deploys the Lambda functions that are needed to retrieve and process the DMARC reports from the S3 bucket and updates the Postgres Database.
1. Run the command to update docker:

   ```shell
   sudo apt install docker.io
   ```
2. Confirm that you have filled in the access key, secret key, default region, and bucket name fields in `installer/env.list`.

3. Choose a username and password for RDS. Enter the values into `./envlist` :
    * The username must follow these rules
        * Must be 1 to 63 letters or numbers.
        * First character must be a letter.
        * Cannot be a reserved word for the chosen database engine.

    * and the password can be from 8-128 characters
```
...
REPORTING_DB_USER=masterUsername
REPORTING_DB_PASSWORD=your-rds-password
...
```
Great! Now you are ready to run the Docker image :) (`REPORTING_DB_HOST` field in `/env.list` can be left empty for now!)

4. Run Docker Image from Docker Hub

*If the following commands prompts permission denied error on your computer, you may need `sudo` privilege.*

You are able to run the install Docker image from the Docker Hub public repository, or you can build it from scratch using our Dockerfile in the `installer` directory.

* Option 1 (Faster): run pre-built Docker image from Docker Hub (navigate to `root` directory):
    ```shell
    docker run -it --env-file env.list --env-file installer/env.list fraudmarc/fraudmarc-ce-install
    ```
* Option 2: build (may take a few minutes) and run Docker image from /installer/Dockerfile (navigate to `root` directory):
    ```shell
    docker build -t fraudmarc-ce-install -f installer/Dockerfile .
    docker run -it --env-file env.list --env-file installer/env.list fraudmarc-ce-install
    ```

If you reached this point, you have successfully created an IAM role, deployed 2 Lambda functions, and launched a RDS database Now let's finish your database configuration and SES setup!

### 3. Configure Your Database:thumbsup:
*If you need maximum security, use a private VPC for your RDS instance and Lambdas. Such configuration is beyond the scope of this document.* 

As mentioned above, your RDS database has already been launched by the Docker image. However, there are some configurations that need to be done.
1. Go to the RDS console -> Instances tab. The database instance that have been created for you can be found as `fraudmarcce`. Click on the instance.

2. Although we have launched the instance, you need to wait a few minutes for it to completely setup (check the `DB instance status` in the Summary panel). 

3. After the DB instance has been successfully created (can be in the `backing up` status), scroll down to the `Connect` panel and copy the value below `Endpoint` (i.e. `fraudmarcce.aaaaaaaaaaaa.aa-aaaa-a.rds.amazonaws.com`) and paste it to `env.list` in the `root` directory as such:

(inside `/env.list`)
```
...
REPORTING_DB_HOST=fraudmarcce.aaaaaaaaaaaa.aa-aaaa-a.rds.amazonaws.com
...
```

4. Go to the AWS Lambda Console and click on the `fraudmarc-ce-process` function. In the `Environment variables` panel below, fill the value for the `REPORTING_DB_HOST` key with the `Endpoint` value . Repeat this step for `fraudmarc-ce-receive`.

5. Install the [PSQL](<https://www.postgresql.org/download/>) command line tool on your local machine.

6. Navigate to the `root` directory and import the Fraudmarc CE schema into your new database with the command (You can find your endpoint on your RDS instance panel) below: 

   ```shell
   cd ..path-to-fraudmarc-ce/
   pg_restore --no-privileges --no-owner -v -h [endpoint of DB instance] -U [DB master username] -n public -d [new DB name (*not instance name*)] fraudmarcce
   ```

7. Go to RDS panel and choose the Instances on the left panel. Choose the `fraudmarcce` instance you just created. Scroll down to `Connect`, and copy the `Endpoint` value to `/env.list` file in the project repository `root` directory like:arrow_down:

    ```
    # Set to match your environment
    REPORTING_DB_NAME=fraudmarcce
    REPORTING_DB_USER=masterUsername
    REPORTING_DB_PASSWORD=your-rds-password
    REPORTING_DB_HOST=your.endpoint
    REPORTING_DB_SSL=require
    REPORTING_DB_MAX_TIME=180s
    ```



### 4. Setup Your AWS Simple Email Service (SES):thumbsup:

1. In AWS SES, choose the Domain on the left side, and click Verify a New Domain. Enter  `fraudmarc-ce.<your domain name>`. Your DMARC Reports will be delievered here. Follow the instructions to setup the Domain Verification Record and Email Receiving Record to complete the verification.

   > If you use the service provider like Godaddy, add a new DNS TXT record with the host field `_amazonses.fraudmarc-ce` (without your domain name) and the TXT field from AWS. Add a new MX record with `fraudmarc-ce` as the host field and the MX field from AWS without the '10'. If your domain service provider requires you to configure the mx record priority, place the '10' in the priority field.

2. Choose the Rule Sets on the left panel. If you are new to AWS, create a Receipt Rule. If you have existing rules, create a new rule.

3. Enter the email address that the DMARC Report will be sent to (`dmarc@fraudmarc-ce.<your domain name>`), and click Next Step.

4. Add action with type S3, and create a S3 bucket with the BUCKET_NAME that you have entered in `/installer/env.list`. *The names must be identical.*

5. Add action with type Lambda and chose your `receive` Lambda function. If prompted, grant permissions.

### 5. Run The Fraudmarc CE Docker:thumbsup:

We've simplified the client side of Fraudmarc CE by providing a single Docker to provide both the Angular frontend and Go backend.

1. As with the Install Docker Image, you are able to run the install Docker image from the Docker Hub public repository, or you can build it from scratch using our Dockerfile in the `root` directory.

* Option 1(Faster): run pre-built Docker image in Docker Hub (navigate to `root` directory):
    ```shell
    docker run -it --env-file env.list --env-file installer/env.list -p 7489:7489 fraudmarc/fraudmarc-ce
    ```
* Option 2: build, Docker image from `/Dockerfile` (the build process may take several minutes):
    * Navigate to the directory containing the Fraudmarc CE docker image (`root` directory).

    * Run the command to download and set up dependencies. This process may take several minutes

    ```shell
    docker build -t fraudmarc-ce .
    ```

    * Run the docker image

    ```shell
    docker run -it --env-file env.list -p 7489:7489 fraudmarc-ce
    ```

2. Your Fraudmarc CE installation is now ready at [http://localhost:7489](http://localhost:7489).

### 6. Create a DMARC policy:thumbsup:

**It may take between a few hours and a few days for your first data to arrive.** 

Be sure you have created a DMARC policy in your domain's DNS so that DMARC aggregate reports will be sent to the address you set in SES step 3 above.

Your DMARC record should be similar to:

   - **Type**: `TXT`
   - **Hostname**: `_dmarc`
   - **Value**: `v=DMARC1; p=none; rua=mailto:dmarc@fraudmarc-ce.<your domain name>; ri=3600;`
       
### :grimacing: Questions?

If you have questions about installation or email authentication, please reference the Fraudmarc Community forum at [https://community.fraudmarc.com](https://community.fraudmarc.com).
