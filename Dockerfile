# - - build
# docker build -t fraudmarc-ce .
# - - run
# docker run -it --env-file env.list -p 7489:7489 fraudmarc-ce
# - - stop & remove all of your docker images in case you wasted a lot of space
# docker stop $(docker ps -a -q); docker rm $(docker ps -a -q); docker rmi -f $(docker images -q); docker images

FROM node:alpine as frontend
COPY /frontend /frontend
WORKDIR /frontend
RUN npm ci --silent && \
    $(npm bin)/ng build --prod --no-progress && \
    cp -R dist / && \
    cd / && \
    rm -rf /frontend /.npm

FROM golang:alpine as backend
COPY /database/dat.patch /
COPY /backend/server /server
RUN apk -Uuv add git upx && \
	rm /var/cache/apk/* && \
    (go get -d gopkg.in/mgutz/dat.v1 ; exit 0) && \
    cd $GOPATH/src/gopkg.in/mgutz/dat.v1 && \
    patch -p1 < /dat.patch && \
    cd /server && \
    go get \
    github.com/fraudmarc/fraudmarc-ce/backend/lib \
    github.com/fraudmarc/fraudmarc-ce/database \
    github.com/gorilla/mux && \
    CGO_ENABLED=0 GOOS=linux \
    go build -a -installsuffix cgo -ldflags '-s -w -extldflags "-static"' -o fraudmarc-ce.sw . && \
    upx -f --brute -o /fraudmarc-ce fraudmarc-ce.sw && \
    cd / && \
    rm -rf $GOPATH /server

FROM scratch
COPY --from=frontend /dist /dist
COPY --from=backend /fraudmarc-ce /server/
WORKDIR /server
CMD ["./fraudmarc-ce"]
EXPOSE 7489
