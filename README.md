# Cloud Foundry Firehose Exporter [![Build Status](https://travis-ci.org/cloudfoundry-community/firehose_exporter.png)](https://travis-ci.org/cloudfoundry-community/firehose_exporter)

A [Prometheus][prometheus] exporter for [Cloud Foundry Firehose][firehose] metrics. It exports `ContainerMetric`, `CounterEvent` and `ValueMetric` events.

## Installation

### Locally

Using the standard `go install` (you must have [Go][golang] already installed in your local machine):

```bash
$ go install github.com/cloudfoundry-community/firehose_exporter
$ firehose_exporter <flags>
```

### Cloud Foundry

The exporter can be deployed to an already existing [Cloud Foundry][cloudfoundry] environment:

```bash
$ git clone https://github.com/cloudfoundry-community/firehose_exporter.git
$ cd firehose_exporter
```

Modify the included [application manifest file][manifest] to include your [Cloud Foundry Firehose][firehose] properties. Then you can push the exporter to your Cloud Foundry environment:

```bash
$ cf push
```

### BOSH

This exporter can be deployed using the [Prometheus BOSH Release][prometheus-boshrelease].

## Usage

### UAA Client

In order to connect to the [Cloud Foundry Firehose][firehose] a `client-id` and `client-secret` must be provided. The `client-id` must have the `doppler.firehose` authority.

For example, to create a new `client-id` and `client-secret` with the right permissions:

```bash
uaac target https://<YOUR UAA URL> --skip-ssl-validation
uaac token client get <YOUR ADMIN CLIENT ID> -s <YOUR ADMIN CLIENT SECRET>
uaac client add prometheus-firehose \
  --name prometheus-firehose \
  --secret prometheus-client-secret \
  --authorized_grant_types client_credentials,refresh_token \
  --authorities doppler.firehose
```

### Flags

| Flag / Environment Variable | Required | Default | Description
| --------------------------- | -------- | ------- | -----------
| uaa.url<br />FIREHOSE_EXPORTER_UAA_URL | Yes | | Cloud Foundry UAA URL
| uaa.client-id<br />FIREHOSE_EXPORTER_UAA_CLIENT_ID | Yes | | Cloud Foundry UAA Client ID
| uaa.client-secret<br />FIREHOSE_EXPORTER_UAA_CLIENT_SECRET | Yes | | Cloud Foundry UAA Client Secret
| doppler.url<br />FIREHOSE_EXPORTER_DOPPLER_URL | Yes | | Cloud Foundry Doppler URL
| doppler.subscription-id<br />FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID | No | prometheus | Cloud Foundry Doppler Subscription ID
| doppler.idle-timeout-seconds<br />FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT_SECONDS | No | 5 | Cloud Foundry Doppler Idle Timeout (in seconds)
| doppler.metric-expiry<br />FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRY | No | 5 minutes | How long a Cloud Foundry Doppler metric is valid
| bosh.deployment<br />FIREHOSE_EXPORTER_DOPPLER_DEPLOYMENTS | No | | Filter metrics to an specific BOSH deployment (this flag can be specified multiple times)
| skip-ssl-verify<br />FIREHOSE_EXPORTER_SKIP_SSL_VERIFY | No | false | Disable SSL Verify |
| metrics.namespace<br />FIREHOSE_EXPORTER_METRICS_NAMESPACE | No | firehose_exporter | Metrics Namespace
| metrics.garbage<br />FIREHOSE_EXPORTER_METRICS_GARBAGE | No | 2 minute | How long to run the metrics garbage
| web.listen-address<br />FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS | No | :9186 | Address to listen on for web interface and telemetry
| web.telemetry-path<br />FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH | No | /metrics | Path under which to expose Prometheus metrics

[cloudfoundry]: https://www.cloudfoundry.org/
[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[golang]: https://golang.org/
[manifest]: https://github.com/cloudfoundry-community/firehose_exporter/blob/master/manifest.yml
[prometheus]: https://prometheus.io/
[prometheus-boshrelease]: https://github.com/cloudfoundry-community/prometheus-boshrelease
