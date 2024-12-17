# frozen_string_literal: true

require_relative 'otel_snitch/version'
require 'otel_snitch/otel_snitch'

module OtelSnitch
  class Error < StandardError; end

  class Exporter
    def initialize(endpoint)
      @endpoint = endpoint
    end

    def write(spans, dir = '/dev/shm/', name = nil)
      name ||= "otel-snitch-file-#{Process.pid}-#{rand(1000)}"
      write_spans(dir + name, dir + "#{name}"+ "-attributes", spans)

      name
    end

    def export(spans, dir = nil, name = nil)
      file = write(spans, dir, name)

      uri = URI.parse("#{@endpoint}?file=#{file}")
      http = Net::HTTP.new(uri.host, uri.port)
      request = Net::HTTP::Post.new(uri.request_uri)

      response = http.request(request)
      response.body
    end

    def span_kind_to_int(span)
      case span.kind
      when 'internal'
        0
      when 'server'
        1
      when 'client'
        2
      when 'producer'
        3
      when 'consumer'
        4
      else
        0
      end
    end

    def status_code(span)
      span.status.code
    end

    def status_description(span)
      span.status.description
    end

    def resource_attributes(span)
      span.resource['attributes']
    end

    def span_id(span)
      span.context.span_id
    end

    def trace_id(span)
      span.context.trace_id
    end

    def trace_flags(span)
      span.context.trace_flags.sampled? ? 1 : 0
    end

    def tracestate_str(span)
      span.context.tracestate.to_s
    end
  end
end
