# frozen_string_literal: true

require "pathname"

module Sunburst
  # Process-wide settings: credential resolution, endpoint, polling and output
  # location. Everything has a sane default so `Sunburst.generate` works with
  # nothing but a prompt.
  class Configuration
    # Where we look for a key, in order. The dotfile keeps the key out of the
    # shell history and out of this repo.
    KEY_ENV_VARS = %w[FAL_KEY FAL_AI_API_KEY FAL_AI_KEY].freeze
    KEY_FILES = ["~/.fal_ai_api_key", "~/.fal_key"].freeze

    DEFAULT_HOST = "https://queue.fal.run"
    DEFAULT_MODEL = "openai/gpt-image-2.5/sunburst/text-to-image"

    # fal queues requests per *app*, not per endpoint variant: you submit to
    # openai/gpt-image-2.5/sunburst/text-to-image but poll
    # openai/gpt-image-2.5/requests/{id}. Submitting returns the exact URLs, and
    # Job prefers those; this is the fallback for looking up a request id later.
    QUEUE_BASE_SEGMENTS = 2

    attr_accessor :host, :model, :poll_interval, :max_wait, :connect_timeout,
                  :read_timeout, :logger, :key_files
    attr_writer :queue_base
    attr_writer :api_key, :output_dir

    def initialize
      @host = ENV.fetch("FAL_HOST", DEFAULT_HOST)
      @model = ENV.fetch("SUNBURST_MODEL", DEFAULT_MODEL)
      @poll_interval = Float(ENV.fetch("SUNBURST_POLL_INTERVAL", 3))
      # Sunburst trades latency for fidelity; max/xhigh 4K frames can take minutes.
      @max_wait = Float(ENV.fetch("SUNBURST_MAX_WAIT", 900))
      @connect_timeout = 15
      @read_timeout = 120
      @output_dir = ENV["SUNBURST_OUTPUT_DIR"]
      @logger = nil
      @key_files = KEY_FILES.dup
      @queue_base = ENV["SUNBURST_QUEUE_BASE"]
    end

    def output_dir
      Pathname.new(@output_dir || File.join(Dir.pwd, "output"))
    end

    # Resolved lazily so merely loading the library never needs a key.
    def api_key
      @api_key ||= api_key_from_env || api_key_from_file ||
                   raise(MissingCredentialsError, missing_key_message)
    end

    def api_key?
      !(api_key_from_env || api_key_from_file).nil?
    rescue StandardError
      false
    end

    # fal's queue API authenticates with `Authorization: Key <token>`.
    def authorization_header
      "Key #{api_key}"
    end

    def endpoint_url
      "#{host}/#{model}"
    end

    # "openai/gpt-image-2.5" for the default model.
    def queue_base
      @queue_base || model.split("/").first(QUEUE_BASE_SEGMENTS).join("/")
    end

    def requests_url(request_id)
      "#{host}/#{queue_base}/requests/#{request_id}"
    end

    def status_url(request_id)
      "#{requests_url(request_id)}/status"
    end

    def cancel_url(request_id)
      "#{requests_url(request_id)}/cancel"
    end

    private

    def api_key_from_env
      KEY_ENV_VARS.each do |name|
        value = ENV[name].to_s.strip
        return value unless value.empty?
      end
      nil
    end

    def api_key_from_file
      key_files.each do |path|
        expanded = File.expand_path(path)
        next unless File.readable?(expanded)

        value = File.read(expanded).strip
        return value unless value.empty?
      end
      nil
    end

    def missing_key_message
      "No fal.ai credential found. Set #{KEY_ENV_VARS.first} or write the key " \
        "to #{KEY_FILES.first}."
    end
  end
end
