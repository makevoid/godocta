# frozen_string_literal: true

require "json"
require "excon"

module Sunburst
  # Thin, explicit wrapper over fal's queue API.
  #
  #   POST   /{model}                        -> { request_id, status_url, ... }
  #   GET    /{app}/requests/{id}/status      -> { status, queue_position, logs }
  #   GET    /{app}/requests/{id}             -> the model output
  #   PUT    /{app}/requests/{id}/cancel
  #
  # Note the asymmetry: you submit to the full endpoint path but poll the app
  # path. Callers should pass the `url:` that submit handed back rather than
  # rely on the derived fallback.
  class QueueClient
    SUCCESS = (200..299)
    # Transient conditions worth one more try before giving up.
    RETRYABLE_STATUSES = [408, 429, 500, 502, 503, 504].freeze
    MAX_ATTEMPTS = 3

    def initialize(config = Sunburst.config)
      @config = config
    end

    # Enqueues a generation. Returns the raw submit response, which carries the
    # request_id used for everything after this.
    def submit(request)
      body = post(@config.endpoint_url, request.to_payload)
      unless body["request_id"]
        raise ApiError, "submit response had no request_id: #{body.inspect}"
      end

      body
    end

    def status(request_id, url: nil, logs: true)
      target = url || @config.status_url(request_id)
      target += (target.include?("?") ? "&" : "?") + "logs=1" if logs
      get(target)
    end

    def result(request_id, url: nil)
      get(url || @config.requests_url(request_id))
    end

    def cancel(request_id, url: nil)
      request(:put, url || @config.cancel_url(request_id))
    end

    private

    def get(url)
      request(:get, url)
    end

    def post(url, payload)
      request(:post, url, payload)
    end

    def request(method, url, payload = nil)
      attempt = 0
      begin
        attempt += 1
        response = Excon.send(
          method, url,
          headers: headers(payload),
          body: payload && JSON.generate(payload),
          connect_timeout: @config.connect_timeout,
          read_timeout: @config.read_timeout,
          # fal redirects queue reads to a result host.
          middlewares: Excon.defaults[:middlewares] + [Excon::Middleware::RedirectFollower]
        )
        handle(response, method, url)
      rescue Excon::Error::Timeout, Excon::Error::Socket => e
        retry if attempt < MAX_ATTEMPTS
        raise ApiError, "#{method.to_s.upcase} #{url} failed after #{attempt} attempts: #{e.message}"
      rescue ApiError => e
        retry if RETRYABLE_STATUSES.include?(e.status) && attempt < MAX_ATTEMPTS
        raise
      end
    end

    def headers(payload)
      h = {
        "Authorization" => @config.authorization_header,
        "Accept" => "application/json",
        "User-Agent" => "sunburst-rb/#{Sunburst::VERSION}"
      }
      h["Content-Type"] = "application/json" if payload
      h
    end

    def handle(response, method, url)
      return parse(response.body) if SUCCESS.cover?(response.status)

      raise ApiError.new(
        "#{method.to_s.upcase} #{url} -> #{response.status}: #{truncate(response.body)}",
        status: response.status,
        body: response.body
      )
    end

    def parse(body)
      return {} if body.nil? || body.strip.empty?

      JSON.parse(body)
    rescue JSON::ParserError => e
      raise ApiError, "could not parse response as JSON: #{e.message}"
    end

    def truncate(text, limit = 500)
      text.to_s[0, limit]
    end
  end
end
