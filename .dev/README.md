# Docker compose stack

This directory contains a small Docker compose stack for local development purposes. It provides an instance of the [OTel Collector](https://github.com/ExplorViz/otel-collector), the required ClickHouse database, as well as the [OpenTelemetry Demo](https://github.com/open-telemetry/opentelemetry-demo) to populate the database with example data.

To run the compose stack, ensure Docker is correctly installed on your system, navigate into this directory, and run:

```shell
docker compose up
```

You can then access the demo frontend at http://localhost:18080 to generate telemetry.
