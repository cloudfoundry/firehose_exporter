# FAQ

### What metrics does this exporter report?

The Cloud Foundry Firehose Prometheus Exporter is a proxy for [Cloud Foundry Firehose][firehose] metrics. It exports Cloud Foundry `ContainerMetric`, `CounterEvent`, `HttpStartStop` and `ValueMetric` metrics.

Metrics are cached (with a expirity defined at the `doppler.metric-expiration` command flag). The exporter always emits the last metric received if it has not been expired.

For a list of all [Cloud Foundry Firehose][firehose] metrics check the [Cloud Foundry Component Metrics][cfmetrics] documentation.

#### ContainerMetric metrics

`ContainerMetric` metrics reports resource usage of an application in a container. The exporter emits:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_container_metric_cpu_percentage | Cloud Foundry Firehose container metric: CPU used, on a scale of 0 to 100 | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_memory_bytes | Cloud Foundry Firehose container metric: bytes of memory used | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_disk_bytes | Cloud Foundry Firehose container metric: bytes of disk used | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_memory_bytes_quota | Cloud Foundry Firehose container metric: maximum bytes of memory allocated to container | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|
| *metrics.namespace*_container_metric_disk_bytes_quota | Cloud Foundry Firehose container metric: maximum bytes of disk allocated to container | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `application_id`, `instance_index`|

#### CounterEvent metrics

`CounterEvent` metrics represents a metric counter (`delta` and `total`). The exporter normalizes each *counter_event_name* received from a [Cloud Foundry Firehose][firehose] *origin* and emits:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_counter_event_*origin*_*counter_event_name*_total | Cloud Foundry Firehose '*counter_event_name*' total counter event from '*origin*' | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip` |
| *metrics.namespace*_counter_event_*origin*_*counter_event_name*_delta | Cloud Foundry Firehose '*counter_event_name*' delta counter event from '*origin*' | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip` |

#### HttpStartStop metrics

*The HttpStartStop Collector is a work in progress. It is suitable for experimentation and may not become supported in the future*

An `HttpStartStop` event represents the whole lifecycle of an HTTP request. The exporter summarizes **applications** related HTTP requests from a slidding window (`doppler.metric-expiration` command flag) from the [Cloud Foundry Firehose][firehose] and emits:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_http_start_stop_requests | Cloud Foundry Firehose http start stop requests | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host`, `status_code` |
| *metrics.namespace*_http_start_stop_response_size_bytes | Summary of Cloud Foundry Firehose http start stop request size in bytes | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host`, `quantile` *[1]* |
| *metrics.namespace*_http_start_stop_response_size_bytes_count | Summary of Cloud Foundry Firehose http start stop request size in bytes (number of observations) | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *metrics.namespace*_http_start_stop_response_size_bytes_sum | Summary of Cloud Foundry Firehose http start stop request size in bytes (sum of observations) | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *metrics.namespace*_http_start_stop_last_request_timestamp | Number of seconds since 1970 since last http start stop received from Cloud Foundry Firehose | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *metrics.namespace*_http_start_stop_client_request_duration_seconds | Summary of Cloud Foundry Firehose http start stop client request duration in seconds | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host`, `quantile` *[1]* |
| *metrics.namespace*_http_start_stop_client_request_duration_seconds_count | Summary of Cloud Foundry Firehose http start stop client request duration in seconds (number of observations) | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *metrics.namespace*_http_start_stop_client_request_duration_seconds_sum | Summary of Cloud Foundry Firehose http start stop client request duration in seconds (sum of observations) | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *metrics.namespace*_http_start_stop_server_request_duration_seconds | Summary of Cloud Foundry Firehose http start stop server request duration in seconds | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host`, `quantile` *[1]* |
| *metrics.namespace*_http_start_stop_server_request_duration_seconds_count | Summary of Cloud Foundry Firehose http start stop server request duration in seconds (number of observations) | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |
| *metrics.namespace*_http_start_stop_server_request_duration_seconds_sum | Summary of Cloud Foundry Firehose http start stop server request duration in seconds (sum of observations) | `bosh_deployment`, `application_id`, `instance_id`, `method`, `scheme`, `host` |

*[1]* Summaries reports [quantiles][quantile] (`0.5`, `0.9`, `0.99`)

#### ValueMetric metrics

`ValueMetric` metrics represents the value of a metric at an instant in time. The exporter normalizes each *value_metric_name* received from a [Cloud Foundry Firehose][firehose] *origin* and emits:

| Metric | Description | Labels |
| ------ | ----------- | ------ |
| *metrics.namespace*_value_metric_*origin*_*value_metric_name* | Cloud Foundry Firehose '*value_metric_name*' value metric from '*origin*' | `origin`, `bosh_deployment`, `bosh_job_name`, `bosh_job_id`, `bosh_job_ip`, `unit` |

### How can I filter by a particular Firehose event?

The `filter.events` command flag allows you to filter what event metrics will be reported (if not set, all events will be enabled by default). Possible values are `ContainerMetric`, `CounterEvent`, `HttpStartStop`, `ValueMetric` (or a combination of them).

### How can I filter metrics coming from a particular BOSH deployment?

The `filter.deployments` command flag allows you to filter metrics which origin is a particular BOSH deployment.

### Can I target multiple Cloud Foundry Firehose endpoints with a single exporter instance?

No, this exporter only supports targetting a single [Cloud Foundry Firehose][firehose] endpoint. If you want to get metrics from several endpoints, you will need to use one exporter per endpoint.

### How can I get readeable names for Container Metrics labels, like the application name?

You can combine this exporter with the [Cloud Foundry Prometheus Exporter][cf_exporter], that provides administrative information about `Applications`, `Organizations`, `Services` and `Spaces`.

For example:

```
firehose_container_metric_cpu_percentage
  * on(application_id)
  group_left(application_name, organization_name, space_name)
  cf_application_info
```

The *on* specifies the matching label, in this case, the *application_id*. The *group_left* specifies what labels (*application_name*, *organization_name*, *space_name*) from the right metric (*cf_application_info*) should be merged into the left metric (*firehose_container_metric_cpu_percentage*).

### I have a question but I don't see it answered at this FAQ

We will be glad to address any questions not answered here. Please, just open a [new issue][issues].

[cf_exporter]: https://github.com/cloudfoundry-community/cf_exporter
[cfmetrics]: https://docs.cloudfoundry.org/loggregator/all_metrics.html
[firehose]: https://docs.cloudfoundry.org/loggregator/architecture.html#firehose
[issues]: https://github.com/cloudfoundry-community/firehose_exporter/issues
[quantile]: https://en.wikipedia.org/wiki/Quantile
