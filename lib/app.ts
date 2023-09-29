import * as cdk from "aws-cdk-lib";
import { RagStackFargate } from "./ragFargate";

const app = new cdk.App();
new RagStackFargate(app, "RagStackFargate");

app.synth();
