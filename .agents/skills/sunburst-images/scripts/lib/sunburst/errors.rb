# frozen_string_literal: true

module Sunburst
  # Base class for everything this library raises.
  class Error < StandardError; end

  # No fal.ai credential could be resolved.
  class MissingCredentialsError < Error; end

  # A prompt/input value is outside what the endpoint accepts.
  class InvalidInputError < Error; end

  # The API answered with a non-success status.
  class ApiError < Error
    attr_reader :status, :body

    def initialize(message, status: nil, body: nil)
      @status = status
      @body = body
      super(message)
    end
  end

  # A queued request did not reach COMPLETED within the allotted time.
  class TimeoutError < Error; end

  # A queued request was cancelled or errored out server-side.
  class JobFailedError < Error; end
end
