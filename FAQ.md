# FAQ

## How to fix your dashboards and alarms with new exporter ?

Replace metrics names to retrieve:

- `[namespace]_http_start_stop_requests` => `[namespace]_http_total`
- `[namespace]_http_start_stop_response_size_bytes` => `[namespace]_response_size_bytes`
- `[namespace]_http_start_stop_response_size_bytes_count` => `[namespace]_response_size_bytes_count`
- `[namespace]_http_start_stop_response_size_bytes_sum` => `[namespace]_response_size_bytes_sum`
- `[namespace]_http_start_stop_response_size_bytes_sum` => `[namespace]_response_size_bytes_sum`
- `[namespace]_http_start_stop_server_request_duration_seconds` => `[namespace]_http_duration_seconds_bucket` (metric is
  now a histogram with bucket)
- `[namespace]_http_start_stop_server_request_duration_seconds_count` => `[namespace]_http_duration_seconds_count`
- `[namespace]_http_start_stop_server_request_duration_seconds_sum` => `[namespace]_http_duration_seconds_sum`
- `[namespace]_http_start_stop_last_request_timestamp` => **metric has been removed to avoid too much cpu work for
  exporter for metric not used in default dashboards or alerts**
- `[namespace]_http_start_stop_client_request_duration_seconds_count` => **metric has been removed because it was
  already not reported anymore on app but only on gorouter metric**
- `[namespace]_http_start_stop_client_request_duration_seconds_sum` => **metric has been removed because it was already
  not reported anymore on app but only on gorouter metric**


### What metrics does this exporter report?

The Cloud Foundry Firehose Prometheus Exporter is a proxy for [Cloud Foundry Firehose][firehose] metrics. It exports
Cloud Foundry `ContainerMetric`, `CounterEvent`, `HttpStartStop` and `ValueMetric` metrics.

Metrics are cached (with a expirity defined at the `doppler.metric-expiration` command flag). The exporter always emits
the last metric received if it has not been expired.

For a list of all [Cloud Foundry Firehose][firehose] metrics check the [Cloud Foundry Component Metrics][cfmetrics]
documentation.

#### ContainerMetric metrics

`ContainerMetric` metrics reports resource usage of an application in a container. The exporter emits:

| Metric | New Metric (retro compat disable) | Description | Labels |
| ------ | --------------------------------- | ----------- | ------ |
| *metrics.namespace*_container_metric_cpu_percentage | *metrics.namespace*_cpu | Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100 | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_memory_bytes | *metrics.namespace*_memory | Cloud Foundry Firehose container metric: bytes of memory used | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_disk_bytes | *metrics.namespace*_disk | Cloud Foundry Firehose container metric: bytes of disk used | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_memory_bytes_quota | *metrics.namespace*_memory_quota | Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_disk_bytes_quota | *metrics.namespace*_disk_quota | Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|

**Important**:
- Use `retro_compat.disable` command flag to deactivate retro-compat mode and use new names.

#### CounterEvent metrics

`CounterEvent` metrics represents a metric counter. The exporter normalizes each *counter_event_name* received from a [Cloud Foundry Firehose][firehose] *origin* and emits:

| Metric | New Metric (retro compat disable) | Description | Labels |
| ------ | --------------------------------- | ----------- | ------ |
| *metrics.namespace*_counter_event_*origin*_*counter_event_name*_total | *metrics.namespace*_*counter_event_name*_total | Cloud Foundry Firehose `counter_event_name` total counter event from `origin` | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip` |
| *metrics.namespace*_counter_event_*origin*_*counter_event_name*_delta | *metrics.namespace*_*counter_event_name*_delta | Cloud Foundry Firehose `counter_event_name` delta counter event from `origin` | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip` |

**Important**:
- Delta is not exported by default, enable it by using `retro_compat.enable_delta` command flag
- Use `retro_compat.disable` command flag to deactivate retro-compat mode and use new names.

#### Http metrics

An `HttpStartStop` event represents the whole lifecycle of an HTTP request. The exporter generate metrics from related
HTTP requests passing by gorouters using rollup mechanism from the [Cloud Foundry Firehose][firehose], those metrics
are:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *
metrics.namespace*_http_total | Cloud Foundry Firehose http total requests | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host`, `status_code` |
| *
metrics.namespace*_http_response_size_bytes | Summary of Cloud Foundry Firehose http request size in bytes with quantiles 0.2, 0.5, 0.75, 0.95 | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host`, `quantile` |
| *
metrics.namespace*_http_response_size_bytes_count | Summary of Cloud Foundry Firehose http request size in bytes (number of observations) | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *
metrics.namespace*_http_response_size_bytes_sum | Summary of Cloud Foundry Firehose http request size in bytes (sum of observations) | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *
metrics.namespace*_http_duration_seconds_count | Histogram of Cloud Foundry Firehose http client request duration in seconds (number of observations) | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *
metrics.namespace*_http_duration_seconds_sum | Histogram of Cloud Foundry Firehose http client request duration in seconds (sum of observations) | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *
metrics.namespace*_http_duration_seconds_bucket | Histogram of Cloud Foundry Firehose http client request duration in seconds (bucket of observations) | `environment`, `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |


#### ValueMetric metrics

`ValueMetric` metrics represents the value of a metric at an instant in time. The exporter normalizes each *
value_metric_name* received from a [Cloud Foundry Firehose][firehose] *origin* and emits:

| Metric | New Metric (retro compat disable) | Description | Labels |
| ------ | --------------------------------- | ----------- | ------ |
| *metrics.namespace*_value_metric_*origin*_*value_metric_name* | *metrics.namespace*_*value_metric_name* | Cloud Foundry Firehose '*value_metric_name*' value metric from '*origin*' | `environment`, `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `unit` |

**Important**:
- Use `retro_compat.disable` command flag to deactivate retro-compat mode and use new names.


### How can I filter by a particular Firehose event?

The `filter.events` command flag allows you to filter what event metrics will be reported (if not set, all events will
be enabled by default). Possible values are `ContainerMetric`, `CounterEvent`, `Http`, `ValueMetric` (or a combination
of them).

### How can I filter metrics coming from a particular BOSH deployment?

The `filter.deployments` command flag allows you to filter metrics which origin is a particular BOSH deployment.

### Can I target multiple Cloud Foundry Firehose endpoints with a single exporter instance?

No, this exporter only supports targetting a single [Cloud Foundry Firehose][firehose] endpoint. If you want to get
metrics from several endpoints, you will need to use one exporter per endpoint.

### How can scale this exporter if I get a metric drops?

You can scale the exporter by increasing the number of exporter instances and using the same `doppler.subscription-id`
command flag. If you use the same subscription ID on each instance, the [Firehose][firehose] evenly distributes events
across all instances of the exporter. For example, if you have two exporters with the same subscription ID,
the [Firehose][firehose] sends half of the events to one exporter and half to the other.

For more information, check the [Scaling Nozzles][scaling-nozzles] documentation.

### How can I get readeable names for Container Metrics labels, like the application name?

You can combine this exporter with the [Cloud Foundry Prometheus Exporter][cf_exporter], that provides administrative
information about `Applications`, `Organizations`, `Services` and `Spaces`.

For example:

```
firehose_container_metric_cpu_percentage
  * on(application_id)
  group_left(application_name, organization_name, space_name)
  cf_application_info
```

The *on* specifies the matching label, in this case, the *application_id*. The *group_left* specifies what labels (*
application_name*, *organization_name*, *space_name*) from the right metric (*cf_application_info*) should be merged
into the left metric (*firehose_container_metric_cpu_percentage*).

### I have a question but I don't see it answered at this FAQ

We will be glad to address any questions not answered here. Please, just open a [new issue][issues].

[cf_exporter]: https://github.com/bosh-prometheus/cf_exporter

[cfmetrics]: https://docs.cloudfoundry.org/loggregator/all_metrics.html

[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose

[issues]: https://github.com/cloudfoundry/firehose_exporter/issues

[quantile]: https://en.wikipedia.org/wiki/Quantile

[scaling-nozzles]: https://docs.cloudfoundry.org/loggregator/log-ops-guide.html#scaling-nozzles
