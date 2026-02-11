import * as cdk from "aws-cdk-lib/core";
import * as fs from "fs";
import * as iam from "aws-cdk-lib/aws-iam";
import * as s3 from "aws-cdk-lib/aws-s3";
import * as cloudfront from "aws-cdk-lib/aws-cloudfront";
import * as origins from "aws-cdk-lib/aws-cloudfront-origins";
import { Construct } from "constructs";
// import * as sqs from 'aws-cdk-lib/aws-sqs';

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // The code that defines your stack goes here

    // 新規S3バケットを2つ作成
    const bucket1 = new s3.Bucket(this, "SzerPublicFileBucket", {
      bucketName: "szer-public-file-prod-20260211",
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      versioned: true,
    });
    const bucket2 = new s3.Bucket(this, "SzerVideoBucket", {
      bucketName: "szer-video-prod-20260211",
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
      versioned: true,
    });

    // CloudFront OAC（オリジンアクセスコントロール）作成
    const oac = new cloudfront.CfnOriginAccessControl(this, "SzerOAC", {
      originAccessControlConfig: {
        name: "szer-oac",
        originAccessControlOriginType: "s3",
        signingBehavior: "always",
        signingProtocol: "sigv4",
        description: "OAC for SZer S3",
      },
    });

    const publicKey = new cloudfront.CfnPublicKey(this, "VideoPublicKey", {
      publicKeyConfig: {
        name: "szer-video-public-key",
        encodedKey: fs.readFileSync("secret/public_key.pem", "utf8"),
        comment: "Public key for signed video URLs",
        callerReference: `${Date.now()}`,
      },
    });

    // CloudFront署名付きURL用Key Group（事前に公開鍵をAWSに登録しておく必要あり）
    const keyGroup = new cloudfront.CfnKeyGroup(this, "SzerVideoKeyGroup", {
      keyGroupConfig: {
        name: "szer-video-key-group",
        items: [publicKey.ref],
        comment: "Key group for signed URLs for video content",
      },
    });

    // CloudFrontディストリビューション（L1）をOAC付きで作成
    const cfnDistribution = new cloudfront.CfnDistribution(this, "SzerCfnDistribution", {
      distributionConfig: {
        enabled: true,
        defaultRootObject: "",
        origins: [
          {
            id: "SzerS3Origin",
            domainName: bucket1.bucketRegionalDomainName,
            originAccessControlId: oac.ref,
            s3OriginConfig: {},
          },
          {
            id: "SzerVideoOrigin",
            domainName: bucket2.bucketRegionalDomainName,
            originAccessControlId: oac.ref,
            s3OriginConfig: {},
          },
        ],
        defaultCacheBehavior: {
          targetOriginId: "SzerS3Origin",
          viewerProtocolPolicy: "redirect-to-https",
          allowedMethods: ["GET", "HEAD"],
          cachedMethods: ["GET", "HEAD"],
          forwardedValues: {
            queryString: false,
            cookies: { forward: "none" },
          },
        },
        cacheBehaviors: [
          {
            pathPattern: "video/*",
            targetOriginId: "SzerVideoOrigin",
            viewerProtocolPolicy: "redirect-to-https",
            allowedMethods: ["GET", "HEAD"],
            cachedMethods: ["GET", "HEAD"],
            forwardedValues: {
              queryString: false,
              cookies: { forward: "none" },
            },
            trustedKeyGroups: [keyGroup.ref], // 署名付きURL/クッキー必須
          },
        ],
        priceClass: "PriceClass_100",
        httpVersion: "http2",
        viewerCertificate: { cloudFrontDefaultCertificate: true },
      },
    });

    // szer-public-file-prod-20260211（bucket1）はCloudFront経由のみ許可
    bucket1.addToResourcePolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: ["s3:GetObject"],
        principals: [new iam.ServicePrincipal("cloudfront.amazonaws.com")],
        resources: [bucket1.arnForObjects("*")],
        conditions: {
          StringEquals: {
            "AWS:SourceArn": `arn:aws:cloudfront::${this.account}:distribution/${cfnDistribution.ref}`,
          },
        },
      }),
    );
    // szer-video-prod-20260211（bucket2）もCloudFront経由のみ許可（/video/*パスで署名付きURL配信）
    bucket2.addToResourcePolicy(
      new iam.PolicyStatement({
        effect: iam.Effect.ALLOW,
        actions: ["s3:GetObject"],
        principals: [new iam.ServicePrincipal("cloudfront.amazonaws.com")],
        resources: [bucket2.arnForObjects("*")],
        conditions: {
          StringEquals: {
            "AWS:SourceArn": `arn:aws:cloudfront::${this.account}:distribution/${cfnDistribution.ref}`,
          },
        },
      }),
    );
    // これで動画はCloudFrontの/video/*パスで署名付きURL配信できる！S3直アクセスやパブリックは不可。

    // バケット名を出力して確認
    new cdk.CfnOutput(this, "Bucket1Name", { value: bucket1.bucketName });
    new cdk.CfnOutput(this, "Bucket2Name", { value: bucket2.bucketName });
    new cdk.CfnOutput(this, "CloudFrontDistributionId", { value: cfnDistribution.ref });
    new cdk.CfnOutput(this, "CloudFrontDomainName", { value: cfnDistribution.attrDomainName });
    new cdk.CfnOutput(this, "account", { value: this.account });
  }
}
