# frozen_string_literal: true

module Sunburst
  # One generated image as fal reports it, plus where we saved it locally.
  class Image
    attr_reader :url, :width, :height, :content_type, :file_name, :file_size
    attr_accessor :path

    def initialize(attrs)
      data = stringify(attrs)
      @url = data["url"] or raise ApiError, "image entry has no url: #{attrs.inspect}"
      @width = data["width"]
      @height = data["height"]
      @content_type = data["content_type"]
      @file_name = data["file_name"]
      @file_size = data["file_size"]
      @path = nil
    end

    def extension
      from_name = File.extname(file_name.to_s)
      return from_name.delete_prefix(".") unless from_name.empty?

      case content_type
      when %r{/jpe?g} then "jpg"
      when %r{/webp} then "webp"
      else "png"
      end
    end

    def dimensions
      return nil unless width && height

      "#{width}x#{height}"
    end

    def to_h
      { url: url, width: width, height: height, content_type: content_type,
        file_name: file_name, file_size: file_size, path: path&.to_s }.compact
    end

    private

    def stringify(attrs)
      attrs.to_h { |key, value| [key.to_s, value] }
    end
  end

  # The outcome of one generation: the request that produced it, the images,
  # and the raw response for anything we did not model.
  class Result
    attr_reader :request_id, :images, :raw, :request, :elapsed

    def initialize(request_id:, body:, request: nil, elapsed: nil)
      @request_id = request_id
      @raw = body
      @request = request
      @elapsed = elapsed
      entries = body["images"] || body[:images] ||
                raise(ApiError, "response had no images key: #{body.inspect}")
      @images = entries.map { |entry| Image.new(entry) }
    end

    def first
      images.first
    end

    def paths
      images.map(&:path).compact
    end

    def to_h
      {
        request_id: request_id,
        elapsed_seconds: elapsed&.round(2),
        prompt: request&.prompt&.to_h,
        prompt_text: request&.prompt&.to_s,
        input: request&.to_payload&.except(:prompt),
        images: images.map(&:to_h)
      }.compact
    end
  end
end
