# frozen_string_literal: true

require_relative "sunburst/version"
require_relative "sunburst/errors"
require_relative "sunburst/configuration"
require_relative "sunburst/prompt"
require_relative "sunburst/image_request"
require_relative "sunburst/result"
require_relative "sunburst/queue_client"
require_relative "sunburst/job"
require_relative "sunburst/downloader"
require_relative "sunburst/generator"
require_relative "sunburst/prompt_library"

# Ruby client for OpenAI GPT Image 2.5 "Sunburst" on fal.ai, plus a library of
# prompts tuned for the kinds of assets the model is good at: character sheets,
# UI kits, game HUDs and light-novel backgrounds.
module Sunburst
  class << self
    def config
      @config ||= Configuration.new
    end

    def configure
      yield config
      config
    end

    def reset_config!
      @config = Configuration.new
    end

    def generator
      Generator.new(config: config)
    end

    # Convenience one-shot: Sunburst.generate(prompt: "...", name: "test")
    def generate(**kwargs)
      generator.generate(**kwargs)
    end
  end
end
