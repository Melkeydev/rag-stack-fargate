# Welcome to Rag Stack Fargate Service

This is an example of an end-to-end stack that uses React, AWS and Go for a fully scalable and hosted applicaton
This stacks comes with middleware, protected routes, logging in and registering users to a Dynamo Database.

This stack consists of:

- Vite, React, Tailwind on the Frontend
- Go, AWS SDK, Docker on the Backend
- DynamoDB, Fargate, Application Load Balancer and CloudFront on the infrastructure

Fargate is the deployment of choice. If you want to deploy your code to a serverless AWS Lambda, check out [this repo](https://github.com/Melkeydev/rag-stack-lambda)

## Enabling HTTPS on an ApplicationLoadBalancedFargateService in AWS CDK (Typescript)

Deploying Fargate on AWS using CDK is straight forward, however enabling HTTPS on an ApplicationLoadBalancedFargateService requires a domain name registered with Route53 with a hosted zone.

This example creates a Fargate service with an Application Load Balancer (ALB). The ALB is configured to listen on port 443 and forward requests to the Fargate service on port 80. The ALB is configured to use a self-signed certificate.

## Prerequisites

- AWS CDK and Typescript should be installed on your system.

- AWS credentials should be configured on your system.

- A domain name registered with Route53 with a hosted zone.

## Steps

In the `ragFargate.ts` file, you can use the `Backend` costruct and pass in the props required to enabled HTTPS (this creates a certificate) via CDK:

```typescript
const backend = new Backend(this, "Backend", {
  domainName: "*.example.com",
  aRecordName: "server.example.com",
  hostedZoneId: "YOUR_HOSTED_ZONE_ID",
  hostedZoneName: "example.com",
});
```

- domainName: The domain name for the certificate. This is the domain name that will be used to access the service. For example, if the domain name is example.com, then the certificate domain name is example.com. It is suggested to use a wildcard in the front to allow the aRecordName to be more specific.

- hostedZoneName: The name of the hosted zone in Route53. This is the domain name that was registered with Route53. For example, if the domain name is example.com, then the hosted zone name is example.com.

- hostedZoneId: The ID of the hosted zone in Route53. This is the ID of the hosted zone that was created when the domain name was registered with Route53. You need to register the domain and get the hostedZoneId first through console.

- aRecordName: The name of the A record in Route53. This is the name of the A record that will be created. For example, if the domain name is example.com, then the A record name is example.com.

The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

- `npm run build` compile typescript to js
- `npm run watch` watch for changes and compile
- `npm run test` perform the jest unit tests
- `cdk deploy` deploy this stack to your default AWS account/region
- `cdk diff` compare deployed stack with current state
- `cdk synth` emits the synthesized CloudFormation template

## Clean up

Run the following command to delete the stack.

- `cdk destroy`
