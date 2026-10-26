# frozen_string_literal: true

require "fileutils"
require "excon"

module Sunburst
  # Saves generated images (and a sidecar JSON of the exact prompt + settings)
  # under the configured output directory.
  #
  # The sidecar is the point: a nice frame is worthless if you cannot tell which
  # prompt produced it.
  class Downloader
    def initialize(config = Sunburst.config)
      @config = config
    end

    # Writes every image in the result and sets Image#path. Returns the result.
    def save(result, name:, subdir: nil)
      dir = target_dir(subdir)
      FileUtils.mkdir_p(dir)

      result.images.each_with_index do |image, index|
        suffix = result.images.size > 1 ? format("-%02d", index + 1) : ""
        filename = "#{slug(name)}#{suffix}.#{image.extension}"
        path = dir.join(filename)
        File.binwrite(path, fetch(image.url))
        image.path = path
      end

      write_sidecar(dir, name, result)
      result
    end

    def fetch(url)
      response = Excon.get(
        url,
        connect_timeout: @config.connect_timeout,
        read_timeout: @config.read_timeout,
        middlewares: Excon.defaults[:middlewares] + [Excon::Middleware::RedirectFollower]
      )
      unless (200..299).cover?(response.status)
        raise ApiError.new("download #{url} -> #{response.status}", status: response.status)
      end

      response.body
    end

    def target_dir(subdir)
      subdir ? @config.output_dir.join(subdir.to_s) : @config.output_dir
    end

    def slug(name)
      name.to_s.downcase.gsub(/[^a-z0-9]+/, "-").gsub(/\A-|-\z/, "")
    end

    private

    def write_sidecar(dir, name, result)
      path = dir.join("#{slug(name)}.json")
      File.write(path, JSON.pretty_generate({ name: name.to_s }.merge(result.to_h)))
      path
    end
  end
end
