# FAQ

### What metrics does this exporter report?

The Cloud Foundry Firehose Prometheus Exporter is a proxy for [Cloud Foundry Firehose][firehose] metrics. It exports Cloud Foundry `ContainerMetric`, `CounterEvent` and `ValueMetric` metrics.

For a list of all [Cloud Foundry Firehose][firehose] metrics check the [Cloud Foundry Component Metrics][cfmetrics] documentation.

### How can I filter by a particular Firehose event?

The *filter.events* command flag allows you to filter what event metrics will be reported. Possible values are `ContainerMetric`, `CounterEvent`, `ValueMetric` (or a combination of them).

### How can I filter metrics coming from a particular BOSH deployment?

The *filter.deployments* command flag allows you to filter metrics which origin is a particular BOSH deployment.

### How can I get readeable names for Container Metrics labels, like the application name?

You can combine this exporter with the [Cloud Foundry Prometheus Exporter][cf_exporter], that provides administrative information about `Applications`, `Organizations` and `Spaces`.

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
