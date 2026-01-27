logging {
  level    = "info"
  write_to = [loki.write.local.receiver]
}


tracing {
  sampling_fraction = 0.1
}


otelcol.receiver.otlp "default" {
  grpc {
    endpoint = "0.0.0.0:4317"
  }
  http {
    endpoint = "0.0.0.0:4318"
  }

  output {
    logs   = [otelcol.processor.attributes.logs_labels.input]
    traces = [otelcol.exporter.otlp.tempo.input]
  }
}

otelcol.exporter.otlp "tempo" {
  client {
    endpoint = "tempo:4317"
    tls {
      insecure             = true
      insecure_skip_verify = true
    }
  }
}

otelcol.exporter.loki "default" {
  forward_to = [loki.write.local.receiver]
}

otelcol.processor.batch "default" {
  output {
    logs = [otelcol.exporter.loki.default.input]
  }
}

otelcol.processor.attributes "logs_labels" {
  action {
    key    = "loki.attribute.labels"
    value  = "service.name,trace_id,span_id"
    action = "insert"
  }

  output {
    logs = [otelcol.processor.batch.default.input]
  }
}


loki.write "local" {
  endpoint {
    url = "http://loki:3100/loki/api/v1/push"
  }
}


// test comment
discovery.docker "logs_integrations_docker" {
  host             = "unix:///var/run/docker.sock"
  refresh_interval = "5s"
}

discovery.relabel "logs_integrations_docker" {
  targets = []

  rule {
    target_label = "job"
    replacement  = "integrations/docker"
  }

  rule {
    target_label = "instance"
    replacement  = constants.hostname
  }

  rule {
    source_labels = ["__meta_docker_container_name"]
    regex         = "/(.*)"
    target_label  = "service_name"
  }

  rule {
    source_labels = ["__meta_docker_container_log_stream"]
    target_label  = "stream"
  }
}

loki.source.docker "logs_integrations_docker" {
  host             = "unix:///var/run/docker.sock"
  targets          = discovery.docker.logs_integrations_docker.targets
  forward_to       = [loki.write.local.receiver]
  relabel_rules    = discovery.relabel.logs_integrations_docker.rules
  refresh_interval = "5s"
}
