# Prometheus Cloudwatch Adapter

The cloudwatch adapter is a service which receives metrics through remote_write and sends them to  [AWS Cloudwatch](https://docs.aws.amazon.com/AmazonCloudWatch/latest/monitoring/working_with_metrics.html).

## Building
```
make build
```

## AWS Setup

The environment variable AWS_REGION must be set.

You must set up authentication supported by the golang aws sdk, for example, via [IRSA](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html), environment, node role, or local configuration.

## AWS Policy
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "cloudwatch:PutMetricData",
            "Resource": "*"
        }
    ]
}
```

## Cloudwatch Limits
- 40kb request size
- 200 transactions per seconds
- max 10 labels per metrics (timeseries with more than 10 labels are ignored)
- max 20 samples per request (every write request gets split up into multiple put metrics requests)
- NaN and Inf Values are not supported (samples with the value NaN or Inf are ignored)

## Prometheus Configuration
To configure Prometheus to send samples to cloudwatch, add the following to your prometheus.yml:
```
remote_write:
  - url: "http://prometheus-cloudwatch-adapter:9513/write"
```

## License

Apache 2.0
