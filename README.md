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

| Flag / Environment Variable | Required | Default | Description |
| --------------------------- | -------- | ------- | ----------- |
| uaa.url<br />FIREHOSE_EXPORTER_UAA_URL | Yes | | Cloud Foundry UAA URL |
| uaa.client-id<br />FIREHOSE_EXPORTER_UAA_CLIENT_ID | Yes | | Cloud Foundry UAA Client ID |
| uaa.client-secret<br />FIREHOSE_EXPORTER_UAA_CLIENT_SECRET | Yes | | Cloud Foundry UAA Client Secret |
| doppler.url<br />FIREHOSE_EXPORTER_DOPPLER_URL | Yes | | Cloud Foundry Doppler URL |
| doppler.subscription-id<br />FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID | No | prometheus | Cloud Foundry Doppler Subscription ID |
| doppler.idle-timeout-seconds<br />FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT_SECONDS | No | 5 | Cloud Foundry Doppler Idle Timeout (in seconds) |
| doppler.metric-expiration<br />FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRATION | No | 5 minutes | How long a Cloud Foundry Container Metric is valid |
| doppler.deployments<br />FIREHOSE_EXPORTER_DOPPLER_DEPLOYMENTS | No | | Comma separated deployments to filter |
| doppler.events<br />FIREHOSE_EXPORTER_DOPPLER_EVENTS| No | | Comma separated events to filter (`ContainerMetric`, `CounterEvent`, `ValueMetric`) |
| skip-ssl-verify<br />FIREHOSE_EXPORTER_SKIP_SSL_VERIFY | No | false | Disable SSL Verify |
| metrics.namespace<br />FIREHOSE_EXPORTER_METRICS_NAMESPACE | No | firehose_exporter | Metrics Namespace |
| metrics.cleanup-interval<br />FIREHOSE_EXPORTER_METRICS_CLEANUP_INTERVAL | No | 2 minutes | Metrics clean up interval |
| web.listen-address<br />FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS | No | :9186 | Address to listen on for web interface and telemetry |
| web.telemetry-path<br />FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH | No | /metrics | Path under which to expose Prometheus metrics |

### Metrics

For a list of [Cloud Foundry Firehose][firehose] metrics check the [Cloud Foundry Component Metrics][cfmetrics] documentation.

The exporter returns the following internal metrics:

| Metric | Description |
| ------ | ----------- |
| *namespace*_total_envelopes_received | Total number of envelopes received from Cloud Foundry Firehose |
| *namespace*_last_envelope_received_timestamp | Number of seconds since 1970 since last envelope received from Cloud Foundry Firehose |
| *namespace*_total_metrics_received | Total number of metrics received from Cloud Foundry Firehose |
| *namespace*_last_metric_received_timestamp | Number of seconds since 1970 since last metric received from Cloud Foundry Firehose |
| *namespace*_total_container_metrics_received | Total number of container metrics received from Cloud Foundry Firehose |
| *namespace*_total_container_metrics_processed | Total number of container metrics processed from Cloud Foundry Firehose |
| *namespace*_last_container_metric_received_timestamp | Number of seconds since 1970 since last container metric received from Cloud Foundry Firehose |
| *namespace*_total_counter_events_received | Total number of counter events received from Cloud Foundry Firehose |
| *namespace*_total_counter_events_processed | Total number of counter events processed from Cloud Foundry Firehose |
| *namespace*_last_counter_event_received_timestamp | Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose |
| *namespace*_total_value_metrics_received | Total number of value metrics received from Cloud Foundry Firehose |
| *namespace*_total_value_metrics_processed | Total number of value metrics processed from Cloud Foundry Firehose |
| *namespace*_last_value_metric_received_timestamp | Number of seconds since 1970 since last value metric received from Cloud Foundry Firehose |
| *namespace*_slow_consumer_alert | Nozzle could not keep up with Cloud Foundry Firehose |
| *namespace*_last_slow_consumer_alert_timestamp | Number of seconds since 1970 since last slow consumer alert received from Cloud Foundry Firehose |

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/cloudfoundry-community/firehose_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/cloudfoundry-community/firehose_exporter/blob/master/LICENSE).

[cloudfoundry]: https://www.cloudfoundry.org/
[cfmetrics]: https://docs.cloudfoundry.org/loggregator/all_metrics.html
[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[golang]: https://golang.org/
[manifest]: https://github.com/cloudfoundry-community/firehose_exporter/blob/master/manifest.yml
[prometheus]: https://prometheus.io/
[prometheus-boshrelease]: https://github.com/cloudfoundry-community/prometheus-boshrelease
