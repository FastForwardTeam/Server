# Start from golang base image
FROM golang:alpine as goBuilder

# ENV GO111MODULE=on

# Install git.
# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
RUN go mod download 

# Copy the source from the current directory to the working Directory inside the container 
COPY . .

# Build the Go app
RUN go build -o main ./src/

# Use NPM to build frontend
FROM node as npmBuilder

WORKDIR /app

COPY src/adminpanel .

RUN rm -rf dist
RUN npx parcel build --public-url /admin

# Start a new stage from scratch
FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=goBuilder /app/main .

COPY --from=npmBuilder /app/dist/ ./static/admin

EXPOSE 5000

#Command to run the executable
CMD ["./main"]
