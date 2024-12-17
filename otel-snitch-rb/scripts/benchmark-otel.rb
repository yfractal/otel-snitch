require 'json'
require 'cpu_time'
require 'opentelemetry-exporter-otlp'

endpoint = ENV['OTEL_SNITCH_RECEIVER_ENDPOINT'] || 'http://0.0.0.0:4318/v1/traces'
exporter = OpenTelemetry::Exporter::OTLP::Exporter.new(endpoint: endpoint)

File.open('data/spans.json') do |file|
  span_data = JSON.parse(file.read)

  spans = span_data.map do |span|
    span_limits = OpenTelemetry::SDK::Trace::SpanLimits.new
    span_parent = OpenTelemetry::Trace::Span.new

    resource = OpenTelemetry::SDK::Resources::Resource.create(span['resource']['attributes'])

    data = span['instrumentation_scope']
    instrumentation_scope = OpenTelemetry::SDK::InstrumentationScope.new(data['name'], data['version'])

    otel_span = OpenTelemetry::SDK::Trace::Span.new(
      nil, nil, span_parent,
      span['name'], span['kind'], OpenTelemetry::Trace.generate_span_id,
      span_limits, [], span['attributes'],
      [], span['end_timestamp'] / 1000 / 1000 / 1000, resource, instrumentation_scope
    )
    otel_span.finish(end_timestamp: span['end_timestamp'] / 1000_000_000)

    otel_span.to_span_data
  end

  spans *= 2
  spans = spans[0...60]

  start_time = Time.now
  cpu_time0 = cpu_time

  exporter.export(spans)

  cpu_time1 = cpu_time
  end_time = Time.now

  puts "Time taken: #{(end_time - start_time) * 1000} ms, CPU time: #{(cpu_time1 - cpu_time0) * 1000} ms"
end
