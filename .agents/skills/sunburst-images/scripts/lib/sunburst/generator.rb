# frozen_string_literal: true

module Sunburst
  # The front door: prompt in, saved file out.
  #
  #   Sunburst::Generator.new.generate(
  #     name: "character-sheet",
  #     prompt: Sunburst::PromptLibrary[:character_sheet],
  #     quality: "xhigh"
  #   )
  class Generator
    def initialize(config: Sunburst.config, client: nil, downloader: nil)
      @config = config
      @client = client || QueueClient.new(config)
      @downloader = downloader || Downloader.new(config)
    end

    # Submits, waits, downloads. Returns a Result with Image#path populated
    # (unless save: false).
    def generate(prompt:, name: "sunburst", subdir: nil, save: true, progress: nil, **options)
      job = submit(prompt: prompt, **options)
      log "submitted #{name} -> #{job.request_id}"
      result = job.wait(&progress || default_progress(name))
      log "completed #{name} in #{result.elapsed.round(1)}s"
      save ? @downloader.save(result, name: name, subdir: subdir) : result
    end

    # Enqueues without waiting -- lets a rake task fan out a whole batch first
    # and only then start polling, which matters when each frame takes minutes.
    def submit(prompt:, **options)
      request = ImageRequest.new(prompt: prompt, **options)
      response = @client.submit(request)
      Job.new(
        request_id: response["request_id"],
        request: request,
        client: @client,
        urls: response,
        config: @config
      )
    end

    # Runs several specs concurrently in the queue, then collects them in order.
    # Each spec is a hash of the keyword arguments #generate takes.
    def generate_batch(specs, subdir: nil, on_event: nil)
      jobs = specs.map do |spec|
        options = spec.dup
        name = options.delete(:name) || "sunburst"
        prompt = options.delete(:prompt)
        options.delete(:subdir)
        job = submit(prompt: prompt, **options)
        on_event&.call(:submitted, name: name, job: job)
        [name, job, options]
      end

      jobs.map do |name, job, _options|
        result = job.wait { |state| on_event&.call(:polled, name: name, state: state) }
        saved = @downloader.save(result, name: name, subdir: subdir)
        on_event&.call(:saved, name: name, result: saved)
        [name, saved]
      rescue Error => e
        on_event&.call(:failed, name: name, error: e)
        [name, e]
      end.to_h
    end

    private

    def default_progress(name)
      last = nil
      lambda do |state|
        line = [state["status"], state["queue_position"] && "queue #{state["queue_position"]}"]
               .compact.join(" ")
        return if line == last

        last = line
        log "#{name}: #{line} (#{state["elapsed"].round}s)"
      end
    end

    def log(message)
      @config.logger&.info("[sunburst] #{message}")
    end
  end
end
