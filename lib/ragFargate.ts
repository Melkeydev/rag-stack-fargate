import { CfnOutput } from "aws-cdk-lib";
import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { Backend } from "./fargate";
import { Frontend } from "./frontend";

export class RagStackFargate extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const backend = new Backend(this, "Backend", {
      domainName: "*.example.com",
      aRecordName: "server.example.com",
      hostedZoneId: "YOUR_HOSTED_ZONE_ID",
      hostedZoneName: "example.com",
    });

    const frontend = new Frontend(this, "Frontend", { apiUrl: backend.apiUrl });
    new CfnOutput(this, "DistributionUrl", { value: frontend.distributionUrl });
  }
}
