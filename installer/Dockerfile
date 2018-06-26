# - - build (from parent directory)
# docker build -t fraudmarc-ce-install -f installer/Dockerfile .

# - - run
# docker run -it --env-file env.list --env-file installer/env.list fraudmarc-ce-install

# OR run pre-built image from public repository https://hub.docker.com/r/fraudmarc/fraudmarc-ce-install/:
# docker run -it --env-file env.list --env-file installer/env.list fraudmarc/fraudmarc-ce-install

# - - stop & remove all of your docker images in case you wasted a lot of space
# docker stop $(docker ps -a -q); docker rm $(docker ps -a -q); docker rmi -f $(docker images -q); docker images

FROM golang:alpine as builder
RUN apk -Uuv add git zip && \
	rm /var/cache/apk/*
RUN (go get -d gopkg.in/mgutz/dat.v1 ; exit 0)
COPY /database/dat.patch /
WORKDIR $GOPATH/src/gopkg.in/mgutz/dat.v1
RUN patch -p1 < /dat.patch
COPY /functions /
RUN go get \
    github.com/aws/aws-lambda-go/lambda \
    github.com/aws/aws-sdk-go/service/lambda \
    github.com/fraudmarc/fraudmarc-ce/backend/lib \
    github.com/fraudmarc/fraudmarc-ce/database \
    golang.org/x/text/encoding
WORKDIR /receive
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static"' -o receive .
RUN zip ../fraudmarc-ce-receive.zip ./receive
WORKDIR /process
RUN CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static"' -o process .
RUN zip ../fraudmarc-ce-process.zip ./process

FROM alpine as installer
RUN apk -Uuv add python py-pip jq && \
	pip install awscli && \
	apk --purge -v del py-pip && \
	rm /var/cache/apk/*
COPY --from=builder /fraudmarc-ce-receive.zip /fraudmarc-ce-process.zip /
COPY /installer/lambda-assume-policy.json / 
COPY /installer/inline-policy.json /
CMD aws rds create-db-instance --db-name $REPORTING_DB_NAME --db-instance-identifier $REPORTING_DB_IDENTIFIER \
        --allocated-storage 20 --db-instance-class db.t2.micro --engine postgres --master-username $REPORTING_DB_USER \
        --master-user-password $REPORTING_DB_PASSWORD > /dev/null \
    && echo "Your Database has been launched! Check the AWS RDS Console -> Instances tab (fraudmarcce)" \
    && export AWS_ROLE_ARN=$(aws iam create-role --role-name FraudmarcCE --assume-role-policy-document file:///lambda-assume-policy.json \
    | jq ".Role.Arn" | tr -d "\"") && sleep 6 \
    && aws iam attach-role-policy --policy-arn arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess --role-name FraudmarcCE \
    && aws iam attach-role-policy --policy-arn arn:aws:iam::aws:policy/CloudWatchLogsFullAccess --role-name FraudmarcCE \
    && echo "Your IAM Role has been created! Check the AWS IAM Console (FraudmarcCE)" \
    && export PROCESS_ARN=$(aws lambda create-function \
    --region $AWS_DEFAULT_REGION \
    --function-name fraudmarc-ce-process \
    --memory 128 \
    --timeout 300 \
    --description "Process DMARC Reports with IP Intelligence" \
    --role $AWS_ROLE_ARN \
    --environment Variables="{ \
      ARRTable=$ARRTable, \
      ARTable=$ARTable, \
      BUCKET_NAME=$BUCKET_NAME, \
      DRE_TABLE=$DRE_TABLE, \
      REPORTING_DB_NAME=$REPORTING_DB_NAME, \
      REPORTING_DB_USER=$REPORTING_DB_USER, \
      REPORTING_DB_PASSWORD=$REPORTING_DB_PASSWORD, \
      REPORTING_DB_HOST=$REPORTING_DB_HOST, \
      REPORTING_DB_SSL=$REPORTING_DB_SSL, \
      REPORTING_DB_MAX_TIME=$REPORTING_DB_MAX_TIME \
    }" \
    --runtime go1.x \
    --zip-file fileb://fraudmarc-ce-process.zip \
    --handler process \
    | jq ".FunctionArn") \
    && aws lambda create-function \
        --region $AWS_DEFAULT_REGION \
        --function-name fraudmarc-ce-receive \
        --memory 1536 \
        --timeout 300 \
        --description "Receive DMARC RUA Reports" \
        --role $AWS_ROLE_ARN \
        --environment Variables="{ \
          ARRTable=$ARRTable, \
          ARTable=$ARTable, \
          BUCKET_NAME=$BUCKET_NAME, \
          DRE_TABLE=$DRE_TABLE, \
          ArnLambdaDmarcARResolveBulk=$PROCESS_ARN, \
          REPORTING_DB_NAME=$REPORTING_DB_NAME, \
          REPORTING_DB_USER=$REPORTING_DB_USER, \
          REPORTING_DB_PASSWORD=$REPORTING_DB_PASSWORD, \
          REPORTING_DB_HOST=$REPORTING_DB_HOST, \
          REPORTING_DB_SSL=$REPORTING_DB_SSL, \
          REPORTING_DB_MAX_TIME=$REPORTING_DB_MAX_TIME \
        }" \
        --runtime go1.x \
        --zip-file fileb://fraudmarc-ce-receive.zip \
        --handler receive \
        | jq ".FunctionArn" | xargs -I {} \
        sed -i "s/ARN/$PROCESS_ARN/g" inline-policy.json \
        && echo "Your Lambda functions have been created! Check the AWS Lambda Console (fraudmarc-ce-receive/process)" \
        && aws iam put-role-policy --role-name FraudmarcCE --policy-name invokeProcessLambda --policy-document file:///inline-policy.json \
        && echo "Your inline-policy has been added to the FraudmarcCE role! Check the IAM Console" \
        && echo "Your AWS Role, RDS, Lambdas has been setup. Fraudmarc CE installation complete."