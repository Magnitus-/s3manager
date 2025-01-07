# S3 Manager

[![Go Report Card](https://goreportcard.com/badge/github.com/cloudlena/s3manager)](https://goreportcard.com/report/github.com/cloudlena/s3manager)
[![Build Status](https://github.com/cloudlena/s3manager/actions/workflows/main.yml/badge.svg)](https://github.com/cloudlena/s3manager/actions)

A Web GUI written in Go to manage S3 buckets from any provider.

![Screenshot](https://raw.githubusercontent.com/cloudlena/s3manager/main/screenshot.png)

## Features

- List all buckets in your account
- Create a new bucket
- List all objects in a bucket
- Upload new objects to a bucket
- Download object from a bucket
- Delete an object in a bucket

## Usage

### Configuration

The application can be configured with the following environment variables:

- `ENDPOINT`: The endpoint of your S3 server (defaults to `s3.amazonaws.com`)
- `REGION`: The region of your S3 server (defaults to `""`)
- `ACCESS_KEY_ID`: Your S3 access key ID (required) (works only if `USE_IAM` is `false`)
- `SECRET_ACCESS_KEY`: Your S3 secret access key (required) (works only if `USE_IAM` is `false`)
- `USE_SSL`: Whether your S3 server uses SSL or not (defaults to `true`)
- `SKIP_SSL_VERIFICATION`: Whether the HTTP client should skip SSL verification (defaults to `false`)
- `SIGNATURE_TYPE`: The signature type to be used (defaults to `V4`; valid values are `V2, V4, V4Streaming, Anonymous`)
- `PORT`: The port the s3manager app should listen on (defaults to `8080`)
- `ADDRESS`: The address the s3manager app should listen on (defaults to any address)
- `ALLOW_DELETE`: Enable buttons to delete objects (defaults to `true`)
- `FORCE_DOWNLOAD`: Add response headers for object downloading instead of opening in a new tab (defaults to `true`)
- `LIST_RECURSIVE`: List all objects in buckets recursively (defaults to `false`)
- `USE_IAM`: Use IAM role instead of key pair (defaults to `false`)
- `IAM_ENDPOINT`: Endpoint for IAM role retrieving (Can be blank for AWS)
- `SSE_TYPE`: Specified server side encrpytion (defaults blank) Valid values can be `SSE`, `KMS`, `SSE-C` all others values don't enable the SSE
- `SSE_KEY`: The key needed for SSE method (only for `KMS` and `SSE-C`)
- `TIMEOUT`: The read and write timout in seconds (default to `600` - 10 minutes)
- `SHARED_BUCKETS_PATH`: The path to an object containing a yaml list of buckets the user has access to, but doesn't own and wouldn't otherwise see listed which can happen in some s3 platforms. Useful to add additional buckets the user should see in the bucket list, but can't otherwise see listed as he/she doesn't own them.
- `CA_CERT`: The path to an optional CA certificate that the client should use to validate the s3 server's certificate.

Alternatively, you can use the above configuration parameters in a file called `config.yml`, `config.json` or any of the formats supported by viper: https://github.com/spf13/viper?tab=readme-ov-file#reading-config-files

The file can be present in the running directory or in a `.s3manager` directory under your home directory.

Note that any parameters passed in a configuration file, the configuration should be lower case.

### Build and Run Locally

1.  Run `make build`
1.  Execute the created binary and visit <http://localhost:8080>

### Run Container image

1. Run `docker run -p 8080:8080 -e 'ACCESS_KEY_ID=XXX' -e 'SECRET_ACCESS_KEY=xxx' cloudlena/s3manager`

### Deploy to Kubernetes

You can deploy s3manager to a Kubernetes cluster using the [Helm chart](https://github.com/sergeyshevch/s3manager-helm).

## Development

### Lint Code

1. Run `make lint`

### Run Tests

1.  Run `make test`

### Build Container Image

The image is available on [Docker Hub](https://hub.docker.com/r/cloudlena/s3manager/).

1.  Run `make build-image`

### Run Locally for Testing

There is an example [docker-compose.yml](https://github.com/cloudlena/s3manager/blob/main/docker-compose.yml) file that spins up an S3 service and the s3manager. You can try it by issuing the following command:

```shell
$ docker-compose up
```
