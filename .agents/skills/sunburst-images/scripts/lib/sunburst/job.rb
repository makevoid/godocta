# frozen_string_literal: true

module Sunburst
  # A submitted request being polled to completion.
  #
  # Sunburst is the slow/high-fidelity variant, so this is deliberately patient:
  # it reports queue position and elapsed time through the progress block rather
  # than sitting silent for minutes.
  class Job
    TERMINAL_FAILURES = %w[FAILED ERROR CANCELLED CANCELED].freeze

    attr_reader :request_id, :request, :client, :urls

    # `urls` is the submit response: fal hands back the exact status/response/
    # cancel URLs, which live under the app path rather than the endpoint path.
    def initialize(request_id:, request:, client:, urls: {}, config: Sunburst.config)
      @request_id = request_id
      @request = request
      @client = client
      @urls = (urls || {}).to_h { |k, v| [k.to_s, v] }
      @config = config
    end

    def status
      client.status(request_id, url: urls["status_url"])
    end

    def cancel
      client.cancel(request_id, url: urls["cancel_url"])
    end

    # Blocks until the request reports COMPLETED, then fetches the output.
    # Yields the status hash on each poll so callers can show progress.
    def wait(max_wait: @config.max_wait, poll_interval: @config.poll_interval, &progress)
      started = monotonic
      loop do
        state = status
        progress&.call(state.merge("elapsed" => monotonic - started))

        case state["status"]
        when "COMPLETED"
          return build_result(monotonic - started)
        when *TERMINAL_FAILURES
          raise JobFailedError, "request #{request_id} ended as #{state["status"]}: #{state.inspect}"
        end

        elapsed = monotonic - started
        if elapsed > max_wait
          raise TimeoutError,
                "request #{request_id} still #{state["status"]} after #{elapsed.round}s " \
                "(max_wait #{max_wait.round}s); it stays in the queue, re-check with " \
                "`rake sunburst:status REQUEST_ID=#{request_id}`"
        end

        sleep(poll_interval)
      end
    end

    private

    def build_result(elapsed)
      Result.new(
        request_id: request_id,
        body: client.result(request_id, url: urls["response_url"]),
        request: request,
        elapsed: elapsed
      )
    end

    def monotonic
      Process.clock_gettime(Process::CLOCK_MONOTONIC)
    end
  end
end
