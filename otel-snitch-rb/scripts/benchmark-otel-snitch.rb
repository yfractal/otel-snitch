require 'json'
require 'cpu_time'
require 'opentelemetry-sdk'
require 'otel_snitch'
require 'net/http'
require 'uri'

endpoint = ENV['OTEL_SNITCH_RECEIVER_ENDPOINT']
exporter = OtelSnitch::Exporter.new(endpoint)

File.open('data/spans.json') do |file|
  span_data = JSON.parse(file.read)

  spans = span_data.map do |span|
    span_limits = OpenTelemetry::SDK::Trace::SpanLimits.new
    span_parent = OpenTelemetry::Trace::Span.new

    otel_span = OpenTelemetry::SDK::Trace::Span.new(
      nil, nil, span_parent,
      span['name'], span['kind'], span['parent_span_id'],
      span_limits, [], span['attributes'],
      [], span['start_timestamp'], span['resource'], span['instrumentation_scope']
    )
    otel_span.finish(end_timestamp: span['end_timestamp'])

    otel_span
  end

  start_time = Time.now
  cpu_time0 = cpu_time

  exporter.export(spans, './', 'abc')

  cpu_time1 = cpu_time
  end_time = Time.now

  puts "Time taken: #{(end_time - start_time) * 1000} ms, CPU time: #{(cpu_time1 - cpu_time0) * 1000} ms"
end