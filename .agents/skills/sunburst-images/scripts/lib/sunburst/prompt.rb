# frozen_string_literal: true

module Sunburst
  # A prompt written as labeled blocks rather than a paragraph.
  #
  # fal's GPT Image 2.5 guide recommends one block per concern -- purpose,
  # subject, materials, light, camera/frame, exclusions -- because it makes a
  # single revision a one-block edit instead of a rewrite, and it keeps the
  # model from averaging two instructions that landed in the same sentence.
  #
  #   Prompt.new(
  #     purpose: "Character reference sheet for a 2D action game",
  #     subject: "...",
  #     exclude: "No watermark, no signature"
  #   ).to_s
  #   # => "PURPOSE: Character reference sheet...\nSUBJECT: ...\nEXCLUDE: ..."
  class Prompt
    # Canonical block order. Anything not listed is appended, upcased, in the
    # order it was given.
    BLOCK_ORDER = %i[
      purpose subject layout style materials palette text light camera
      composition technical exclude
    ].freeze

    # A few blocks read better under a different label than their key.
    LABELS = {
      camera: "CAMERA/FRAME",
      text: "TEXT",
      technical: "TECHNICAL"
    }.freeze

    attr_reader :blocks

    def initialize(blocks = {})
      @blocks = {}
      blocks.each { |key, value| self[key] = value }
    end

    # Matches a block header such as "CAMERA/FRAME:" at the start of a line.
    HEADER = /\A([A-Z][A-Z0-9 \/_-]*):\s*(.*)\z/

    # Accepts a String, a Prompt, or a Hash of blocks. A String already written
    # as labeled blocks round-trips back into blocks rather than being buried
    # under SUBJECT -- so `coerce(prompt.to_s)` equals `prompt`.
    def self.coerce(value)
      case value
      when Prompt then value
      when Hash then new(value)
      when String then parse(value)
      else
        raise InvalidInputError, "Cannot build a prompt from #{value.class}"
      end
    end

    # Reads "LABEL: text" lines back into blocks, folding unlabeled
    # continuation lines into the block above them. Text with no labels at all
    # becomes a single SUBJECT block.
    def self.parse(text)
      blocks = {}
      current = nil

      text.to_s.lines.each do |line|
        stripped = line.strip
        next if stripped.empty?

        if (match = HEADER.match(stripped))
          current = key_for(match[1])
          blocks[current] = [match[2]]
        elsif current
          blocks[current] << stripped
        else
          current = :subject
          blocks[current] = [stripped]
        end
      end

      new(blocks)
    end

    def self.key_for(label)
      LABELS.each { |key, text| return key if text == label }
      label.downcase.tr(" /-", "___").squeeze("_").to_sym
    end

    def [](key)
      @blocks[key.to_sym]
    end

    def []=(key, value)
      text = normalize(value)
      if text.nil?
        @blocks.delete(key.to_sym)
      else
        @blocks[key.to_sym] = text
      end
    end

    # Returns a copy with the given blocks replaced -- the "one change per pass"
    # workflow without mutating the library's shared example prompts.
    def merge(overrides)
      self.class.new(@blocks.merge(symbolize(overrides)))
    end

    def empty?
      @blocks.empty?
    end

    def to_s
      ordered_keys.map { |key| "#{label_for(key)}: #{@blocks[key]}" }.join("\n")
    end

    def to_h
      ordered_keys.to_h { |key| [key, @blocks[key]] }
    end

    def inspect
      "#<Sunburst::Prompt #{ordered_keys.join(", ")}>"
    end

    private

    def ordered_keys
      known = BLOCK_ORDER & @blocks.keys
      known + (@blocks.keys - BLOCK_ORDER)
    end

    def label_for(key)
      LABELS.fetch(key) { key.to_s.tr("_", " ").upcase }
    end

    def normalize(value)
      return nil if value.nil?

      text = Array(value).flatten.map { |part| part.to_s.strip }.reject(&:empty?).join(" ")
      text = text.gsub(/\s+/, " ").strip
      text.empty? ? nil : text
    end

    def symbolize(hash)
      hash.to_h { |key, value| [key.to_sym, value] }
    end
  end
end
