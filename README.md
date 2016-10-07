# Cloud Foundry Firehose Exporter [![Build Status](https://travis-ci.org/cloudfoundry-community/firehose_exporter.png)](https://travis-ci.org/cloudfoundry-community/firehose_exporter)

A [Prometheus][prometheus] exporter for [Cloud Foundry Firehose][firehose] metrics. It exports `ContainerMetric`, `CounterEvent` and `ValueMetric` events.

## Building and running

```bash
make
./firehose_exporter <flags>
```

### Flags

| Flag | Environment Variable | Required | Default | Description
| ---- | -------------------- | -------- | ------- | -----------
| metrics.namespace | NOZZLE_METRICS_NAMESPACE | No | firehose_exporter | Metrics Namespace
| metrics.garbage | NOZZLE_METRICS_GARBAGE | No | 1 minute | How long to run the metrics garbage
| web.listen-address | NOZZLE_WEB_LISTEN_ADDRESS | No | :9186 | Address to listen on for web interface and telemetry
| web.telemetry-path | NOZZLE_WEB_TELEMETRY_PATH | No | /metrics | Path under which to expose Prometheus metrics
| uaa.url | NOZZLE_UAA_URL | Yes | | Cloud Foundry UAA URL
| uaa.client-id | NOZZLE_UAA_CLIENT_ID | Yes | | Cloud Foundry UAA Client ID
| uaa.client-secret | NOZZLE_UAA_CLIENT_SECRET | Yes | | Cloud Foundry UAA Client Secret
| doppler.url | NOZZLE_DOPPLER_URL | Yes | | Cloud Foundry Doppler URL
| doppler.subscription-id | NOZZLE_DOPPLER_SUBSCRIPTION_ID | No | prometheus | Cloud Foundry Doppler Subscription ID
| doppler.idle-timeout-seconds| NOZZLE_DOPPLER_IDLE_TIMEOUT_SECONDS | No | 5 | Cloud Foundry Doppler Idle Timeout (in seconds)
| doppler.metric-expiry | NOZZLE_DOPPLER_METRIC_EXPIRY | No | 5 minutes | How long a Cloud Foundry Doppler metric is valid
| skip-ssl-verify | NOZZLE_SKIP_SSL_VERIFY | No | false | Disable SSL Verify |

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

## Running tests

```bash
make test
```

## Using Docker

You can deploy this exporter using the [frodenas/firehose-exporter][hub] Docker image. For example:

```bash
docker pull frodenas/firehose-exporter

docker run -d -p 9186:9186 frodenas/firehose-exporter <flags>
```

[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[hub]: https://hub.docker.com/r/frodenas/firehose-exporter/
[prometheus]: https://prometheus.io/
