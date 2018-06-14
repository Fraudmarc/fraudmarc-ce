# - - build (from parent directory)
# docker build -t fraudmarc-ce-install -f installer/Dockerfile .
# - - run
# docker run -it --env-file env.list --env-file installer/env.list fraudmarc-ce-install
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
RUN GOOS=linux go build
RUN zip ../fraudmarc-ce-receive.zip ./receive
WORKDIR /process
RUN GOOS=linux go build
RUN zip ../fraudmarc-ce-process.zip ./process

FROM alpine as installer
RUN apk -Uuv add python py-pip jq && \
	pip install awscli && \
	apk --purge -v del py-pip && \
	rm /var/cache/apk/*
COPY --from=builder /fraudmarc-ce-receive.zip /fraudmarc-ce-process.zip /
CMD aws lambda create-function \
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
    | jq ".FunctionArn" | xargs -I {} \
    aws lambda create-function \
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
          ArnLambdaDmarcARResolveBulk={}, \
          REPORTING_DB_NAME=$REPORTING_DB_NAME, \
          REPORTING_DB_USER=$REPORTING_DB_USER, \
          REPORTING_DB_PASSWORD=$REPORTING_DB_PASSWORD, \
          REPORTING_DB_HOST=$REPORTING_DB_HOST, \
          REPORTING_DB_SSL=$REPORTING_DB_SSL, \
          REPORTING_DB_MAX_TIME=$REPORTING_DB_MAX_TIME \
        }" \
        --runtime go1.x \
        --zip-file fileb://fraudmarc-ce-receive.zip \
        --handler receive
