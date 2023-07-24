## Build
##
FROM golang:1.20-alpine as dev-env

# Copy application data into image
COPY . /Users/mishashevnuk/GolandProjects/app2.4
WORKDIR /Users/mishashevnuk/GolandProjects/app2.4


COPY . .
COPY .env .env

RUN go mod download

# Copy only .go files, if you want all files to be copied then replace with `COPY . . for the code below.

# Build our application.
# RUN go build -o /go/src/bartmika/mullberry-backend/bin/mullberry-backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -gcflags "all=-N -l" -o /server

##
## Deploy
##
FROM alpine:latest
RUN mkdir /data

COPY --from=dev-env /server ./

COPY .env .env
CMD ["./server"]

