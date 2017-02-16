# Cloud Foundry Firehose Exporter [![Build Status](https://travis-ci.org/cloudfoundry-community/firehose_exporter.png)](https://travis-ci.org/cloudfoundry-community/firehose_exporter)

A [Prometheus][prometheus] exporter proxy for [Cloud Foundry Firehose][firehose] metrics. Please refer to the [FAQ][faq] for general questions about this exporter.

## Architecture overview

![](https://cdn.rawgit.com/cloudfoundry-community/firehose_exporter/master/architecture/architecture.svg)

## Installation

### Binaries

Download the already existing [binaries](https://github.com/cloudfoundry-community/firehose_exporter/releases) for your platform:

```bash
$ ./firehose_exporter <flags>
```

### From source

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
| `uaa.url`<br />`FIREHOSE_EXPORTER_UAA_URL` | Yes | | Cloud Foundry UAA URL |
| `uaa.client-id`<br />`FIREHOSE_EXPORTER_UAA_CLIENT_ID` | Yes | | Cloud Foundry UAA Client ID |
| `uaa.client-secret`<br />`FIREHOSE_EXPORTER_UAA_CLIENT_SECRET` | Yes | | Cloud Foundry UAA Client Secret |
| `doppler.url`<br />`FIREHOSE_EXPORTER_DOPPLER_URL` | Yes | | Cloud Foundry Doppler URL |
| `doppler.subscription-id`<br />`FIREHOSE_EXPORTER_DOPPLER_SUBSCRIPTION_ID` | No | `prometheus` | Cloud Foundry Doppler Subscription ID |
| `doppler.idle-timeout`<br />`FIREHOSE_EXPORTER_DOPPLER_IDLE_TIMEOUT` | No | | Cloud Foundry Doppler Idle Timeout duration |
| `doppler.min-retry-delay`<br />`FIREHOSE_EXPORTER_DOPPLER_MIN_RETRY_DELAY` | No | | Cloud Foundry Doppler min retry delay duration |
| `doppler.max-retry-delay`<br />`FIREHOSE_EXPORTER_DOPPLER_MAX_RETRY_DELAY` | No | | Cloud Foundry Doppler max retry delay duration |
| `doppler.metric-expiration`<br />`FIREHOSE_EXPORTER_DOPPLER_METRIC_EXPIRATION` | No | `5 minutes` | How long Cloud Foundry metrics received from the Firehose are valid |
| `filter.deployments`<br />`FIREHOSE_EXPORTER_FILTER_DEPLOYMENTS` | No | | Comma separated deployments to filter |
| `filter.events`<br />`FIREHOSE_EXPORTER_FILTER_EVENTS` | No | | Comma separated events to filter. If not set, all events will be enabled (`ContainerMetric`, `CounterEvent`, `HttpStartStop`, `ValueMetric`) |
| `metrics.namespace`<br />`FIREHOSE_EXPORTER_METRICS_NAMESPACE` | No | `firehose` | Metrics Namespace |
| `metrics.environment`<br />`FIREHOSE_EXPORTER_METRICS_ENVIRONMENT` | No | | Environment label to be attached to metrics |
| `metrics.cleanup-interval`<br />`FIREHOSE_EXPORTER_METRICS_CLEANUP_INTERVAL` | No | `2 minutes` | Metrics clean up interval |
| `skip-ssl-verify`<br />`FIREHOSE_EXPORTER_SKIP_SSL_VERIFY` | No | `false` | Disable SSL Verify |
| `web.listen-address`<br />`FIREHOSE_EXPORTER_WEB_LISTEN_ADDRESS` | No | `:9186` | Address to listen on for web interface and telemetry |
| `web.telemetry-path`<br />`FIREHOSE_EXPORTER_WEB_TELEMETRY_PATH` | No | `/metrics` | Path under which to expose Prometheus metrics |
| `web.auth.username`<br />`FIREHOSE_EXPORTER_WEB_AUTH_USERNAME` | No | | Username for web interface basic auth |
| `web.auth.pasword`<br />`FIREHOSE_EXPORTER_WEB_AUTH_PASSWORD` | No | | Password for web interface basic auth |
| `web.tls.cert_file`<br />`FIREHOSE_EXPORTER_WEB_TLS_CERTFILE` | No | | Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate |
| `web.tls.key_file`<br />`FIREHOSE_EXPORTER_WEB_TLS_KEYFILE` | No | | Path to a file that contains the TLS private key (PEM format) |

### Metrics

For a list of [Cloud Foundry Firehose][firehose] metrics check the [Cloud Foundry Component Metrics][cfmetrics] documentation.

The exporter returns additionally the following internal metrics:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_total_envelopes_received | Total number of envelopes received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_envelope_received_timestamp | Number of seconds since 1970 since last envelope received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_metrics_received | Total number of metrics received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_metric_received_timestamp | Number of seconds since 1970 since last metric received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_container_metrics_received | Total number of container metrics received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_container_metrics_processed | Total number of container metrics processed from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_container_metrics_cached | Number of container metrics cached from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_container_metric_received_timestamp | Number of seconds since 1970 since last container metric received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_counter_events_received | Total number of counter events received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_counter_events_processed | Total number of counter events processed from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_counter_events_cached | Number of counter events cached from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_counter_event_received_timestamp | Number of seconds since 1970 since last counter event received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_http_start_stop_received | Total number of http start stop received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_http_start_stop_processed | Total number of http start stop processed from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_http_start_stop_cached | Number of http start stop cached from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_http_start_stop_received_timestamp | Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_value_metrics_received | Total number of value metrics received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_total_value_metrics_processed | Total number of value metrics processed from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_value_metrics_cached | Number of value metrics cached from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_value_metric_received_timestamp | Number of seconds since 1970 since last value metric received from Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_slow_consumer_alert | Nozzle could not keep up with Cloud Foundry Firehose | `environment` |
| *metrics.namespace*_last_slow_consumer_alert_timestamp | Number of seconds since 1970 since last slow consumer alert received from Cloud Foundry Firehose | `environment` |

## Contributing

Refer to [CONTRIBUTING.md](https://github.com/cloudfoundry-community/firehose_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/cloudfoundry-community/firehose_exporter/blob/master/LICENSE).

[cloudfoundry]: https://www.cloudfoundry.org/
[cfmetrics]: https://docs.cloudfoundry.org/loggregator/all_metrics.html
[faq]: https://github.com/cloudfoundry-community/firehose_exporter/blob/master/FAQ.md
[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[golang]: https://golang.org/
[manifest]: https://github.com/cloudfoundry-community/firehose_exporter/blob/master/manifest.yml
[prometheus]: https://prometheus.io/
[prometheus-boshrelease]: https://github.com/cloudfoundry-community/prometheus-boshrelease
