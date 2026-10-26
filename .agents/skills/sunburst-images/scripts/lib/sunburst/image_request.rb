# frozen_string_literal: true

module Sunburst
  # Validated input payload for openai/gpt-image-2.5/sunburst/text-to-image.
  #
  # Validation happens locally so a bad enum costs nothing instead of a queued
  # request plus a round trip.
  class ImageRequest
    SIZE_PRESETS = %w[
      square_hd square portrait_4_3 portrait_16_9 landscape_4_3 landscape_16_9 auto
    ].freeze

    QUALITIES = %w[auto low medium high xhigh max].freeze
    BACKGROUNDS = %w[auto transparent opaque].freeze
    FORMATS = %w[jpeg png webp].freeze

    # Custom sizes must have both edges divisible by this, with the long edge
    # capped and the aspect ratio no wider than MAX_ASPECT.
    SIZE_MULTIPLE = 16
    MAX_EDGE = 3840
    MIN_EDGE = 16
    MAX_ASPECT = 3.0

    # Formats that can carry an alpha channel.
    ALPHA_FORMATS = %w[png webp].freeze
    # Compression only applies to lossy/compressible formats.
    COMPRESSIBLE_FORMATS = %w[jpeg webp].freeze

    attr_reader :prompt, :image_size, :quality, :background, :output_format,
                :output_compression, :num_images, :sync_mode

    def initialize(prompt:, image_size: "landscape_4_3", quality: "high",
                   background: "auto", output_format: "png",
                   output_compression: nil, num_images: 1, sync_mode: false)
      @prompt = Prompt.coerce(prompt)
      raise InvalidInputError, "prompt is empty" if @prompt.empty?

      @image_size = validate_size(image_size)
      @quality = validate_enum(:quality, quality, QUALITIES)
      @background = validate_enum(:background, background, BACKGROUNDS)
      @output_format = validate_enum(:output_format, output_format, FORMATS)
      @num_images = validate_num_images(num_images)
      @output_compression = validate_compression(output_compression)
      @sync_mode = !!sync_mode

      validate_background_format!
    end

    # The JSON body fal expects. Nils are dropped so the server's own defaults
    # apply rather than us guessing them.
    def to_payload
      {
        prompt: prompt.to_s,
        image_size: image_size,
        quality: quality,
        background: background,
        output_format: output_format,
        output_compression: output_compression,
        num_images: num_images,
        sync_mode: sync_mode
      }.compact
    end

    # Rough label used for output filenames and the run log.
    def size_label
      case image_size
      when String then image_size
      else "#{image_size[:width]}x#{image_size[:height]}"
      end
    end

    private

    def validate_enum(name, value, allowed)
      text = value.to_s
      return text if allowed.include?(text)

      raise InvalidInputError,
            "#{name} must be one of #{allowed.join(", ")} (got #{value.inspect})"
    end

    def validate_size(value)
      case value
      when String, Symbol
        validate_enum(:image_size, value, SIZE_PRESETS)
      when Hash
        validate_custom_size(value)
      when Array
        validate_custom_size(width: value[0], height: value[1])
      else
        raise InvalidInputError, "image_size must be a preset name or {width:, height:}"
      end
    end

    def validate_custom_size(hash)
      dims = hash.to_h { |k, v| [k.to_sym, v] }
      width = Integer(dims.fetch(:width))
      height = Integer(dims.fetch(:height))

      [["width", width], ["height", height]].each do |name, edge|
        unless (edge % SIZE_MULTIPLE).zero?
          raise InvalidInputError, "#{name} #{edge} must be divisible by #{SIZE_MULTIPLE}"
        end
        unless edge.between?(MIN_EDGE, MAX_EDGE)
          raise InvalidInputError, "#{name} #{edge} must be between #{MIN_EDGE} and #{MAX_EDGE}"
        end
      end

      aspect = [width, height].max.to_f / [width, height].min
      if aspect > MAX_ASPECT
        raise InvalidInputError,
              format("aspect ratio %.2f:1 exceeds the %g:1 limit", aspect, MAX_ASPECT)
      end

      { width: width, height: height }
    rescue KeyError => e
      raise InvalidInputError, "image_size hash needs both :width and :height (#{e.message})"
    rescue TypeError, ArgumentError
      raise InvalidInputError, "image_size width/height must be integers"
    end

    def validate_num_images(value)
      count = Integer(value)
      raise InvalidInputError, "num_images must be >= 1" if count < 1

      count
    rescue TypeError, ArgumentError
      raise InvalidInputError, "num_images must be an integer"
    end

    def validate_compression(value)
      return nil if value.nil?

      unless COMPRESSIBLE_FORMATS.include?(@output_format)
        raise InvalidInputError,
              "output_compression only applies to #{COMPRESSIBLE_FORMATS.join("/")}, " \
              "not #{@output_format}"
      end

      level = Integer(value)
      raise InvalidInputError, "output_compression must be 0..100" unless level.between?(0, 100)

      level
    rescue TypeError, ArgumentError
      raise InvalidInputError, "output_compression must be an integer 0..100"
    end

    def validate_background_format!
      return unless @background == "transparent"
      return if ALPHA_FORMATS.include?(@output_format)

      raise InvalidInputError,
            "transparent background requires #{ALPHA_FORMATS.join(" or ")} output_format"
    end
  end
end
