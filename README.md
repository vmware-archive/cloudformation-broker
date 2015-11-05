# AWS CloudFormation Service Broker [![Build Status](https://travis-ci.org/cf-platform-eng/cloudformation-broker.png)](https://travis-ci.org/cf-platform-eng/cloudformation-broker)

This is an **experimental** [Cloud Foundry Service Broker](https://docs.cloudfoundry.org/services/overview.html) for [Amazon CloudFormation](https://aws.amazon.com/cloudformation/).

## Disclaimer

This is **NOT** presently a production ready Service Broker. This is a work in progress. It is suitable for experimentation and may not become supported in the future.

## Installation

### Locally

Using the standard `go install` (you must have [Go](https://golang.org/) already installed in your local machine):

```
$ go install github.com/cf-platform-eng/cloudformation-broker
$ cloudformation-broker -port=3000 -config=<path-to-your-config-file>
```

### Cloud Foundry

The broker can be deployed to an already existing [Cloud Foundry](https://www.cloudfoundry.org/) installation:

```
$ git clone https://github.com/cf-platform-eng/cloudformation-broker.git
$ cd cloudformation-broker
```

Modify the [included manifest file](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/manifest.yml) to include your AWS credentials and the [sample configuration file](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/config-sample.json) to add/update the AWS CloudFormation template URL. Then you can push the broker to your [Cloud Foundry](https://www.cloudfoundry.org/) environment:

```
$ cp config-sample.json config.json
$ cf push cloudformation-broker
```

### Docker

If you want to run the AWS CloudFormation Service Broker on a Docker container, you can use the [cfplatformeng/cloudformation-broker](https://registry.hub.docker.com/u/cfplatformeng/cloudformation-broker/) Docker image.

```
$ docker run -d --name cloudformation-broker -p 3000:3000 \
  -e AWS_ACCESS_KEY_ID=<your-aws-access-key-id> \
  -e AWS_SECRET_ACCESS_KEY=<your-aws-secret-access-key> \
  cfplatformeng/cloudformation-broker
```

The Docker image cames with an [embedded sample configuration file](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/config-sample.json). If you want to override it and/or modify the sample AWS CloudFormation template to use, you can create the Docker image with you custom configuration file by running:

```
$ git clone https://github.com/cf-platform-eng/cloudformation-broker.git
$ cd cloudformation-broker
$ bin/build-docker-image
```

### BOSH

This broker can be deployed using the [AWS Service Broker BOSH Release](https://github.com/cf-platform-eng/aws-broker-boshrelease).

## Configuration

Refer to the [Configuration](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md) instructions.

## Usage

### Managing Service Broker

Configure and deploy the broker using one of the above methods. Then:

1. Check that your Cloud Foundry installation supports [Service Broker API Version v2.6 or greater](https://docs.cloudfoundry.org/services/api.html#changelog)
2. [Register the broker](https://docs.cloudfoundry.org/services/managing-service-brokers.html#register-broker) within your Cloud Foundry installation;
3. [Make Services and Plans public](https://docs.cloudfoundry.org/services/access-control.html#enable-access);
4. Depending on your Cloud Foundry settings, you migh also need to create/bind an [Application Security Group](https://docs.cloudfoundry.org/adminguide/app-sec-groups.html) to allow access to the AWS Resources created by the AWS CloudFormation Stack.

### Integrating Service Instances with Applications

Application Developers can start to consume the services using the standard [CF CLI commands](https://docs.cloudfoundry.org/devguide/services/managing-services.html).

Depending on the [broker configuration](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/CONFIGURATION.md#cloudformation-broker-configuration), Application Depevelopers can send arbitrary parameters on certain broker calls:

#### Provision

Provision calls support optional [arbitrary parameters](https://docs.cloudfoundry.org/devguide/services/managing-services.html#arbitrary-params-create). These parameters will be passed to the CloudFormation Stack as input parameters.

#### Update

Update calls support optional [arbitrary parameters](https://docs.cloudfoundry.org/devguide/services/managing-services.html#arbitrary-params-update). These parameters will be passed to the CloudFormation Stack as input parameters.

## Contributing

In the spirit of [free software](http://www.fsf.org/licensing/essays/free-sw.html), **everyone** is encouraged to help improve this project.

Here are some ways *you* can contribute:

* by using alpha, beta, and prerelease versions
* by reporting bugs
* by suggesting new features
* by writing or editing documentation
* by writing specifications
* by writing code (**no patch is too small**: fix typos, add comments, clean up inconsistent whitespace)
* by refactoring code
* by closing [issues](https://github.com/cf-platform-eng/cloudformation-broker/issues)
* by reviewing patches

### Submitting an Issue

We use the [GitHub issue tracker](https://github.com/cf-platform-eng/cloudformation-broker/issues) to track bugs and features. Before submitting a bug report or feature request, check to make sure it hasn't already been submitted. You can indicate support for an existing issue by voting it up. When submitting a bug report, please include a [Gist](http://gist.github.com/) that includes a stack trace and any details that may be necessary to reproduce the bug, including your Golang version and operating system. Ideally, a bug report should include a pull request with failing specs.

### Submitting a Pull Request

1. Fork the project.
2. Create a topic branch.
3. Implement your feature or bug fix.
4. Commit and push your changes.
5. Submit a pull request.

## Copyright

Copyright (c) 2015 Pivotal Software Inc. See [LICENSE](https://github.com/cf-platform-eng/cloudformation-broker/blob/master/LICENSE) for details.
