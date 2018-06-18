# - - build
# docker build -t fraudmarc-ce .
# - - run
# docker run -it --env-file env.list -p 7489:7489 fraudmarc-ce
# - - stop & remove all of your docker images in case you wasted a lot of space
# docker stop $(docker ps -a -q); docker rm $(docker ps -a -q); docker rmi -f $(docker images -q); docker images

FROM node:alpine as frontend
COPY /frontend/package.json /frontend/package-lock.json ./
RUN npm i && mkdir /frontend && cp -R ./node_modules ./frontend
WORKDIR /frontend
COPY /frontend .
RUN $(npm bin)/ng build --prod

FROM golang:alpine as backend
RUN apk -Uuv add git && \
	rm /var/cache/apk/*
RUN (go get -d gopkg.in/mgutz/dat.v1 ; exit 0)
COPY /database/dat.patch /
WORKDIR $GOPATH/src/gopkg.in/mgutz/dat.v1
RUN patch -p1 < /dat.patch
COPY /backend/server /server
WORKDIR /server
RUN go get \
    github.com/fraudmarc/fraudmarc-ce/backend/lib \
    github.com/fraudmarc/fraudmarc-ce/database \
    github.com/gorilla/mux
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o fraudmarc-ce .

FROM scratch
COPY --from=frontend /frontend/dist /dist
COPY --from=backend /server/fraudmarc-ce /server/
WORKDIR /server
CMD ["./fraudmarc-ce"]
EXPOSE 7489
