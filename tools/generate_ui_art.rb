#!/usr/bin/env ruby
# Regenerate the two Sunburst art assets; keeps exact request sidecars beside them.
require "logger"
require "pathname"
root = Pathname.new(__dir__).parent
$LOAD_PATH.unshift(root.join(".agents/skills/sunburst-images/scripts/lib").to_s)
require "sunburst"

Sunburst.configure do |config|
  config.output_dir = root.join("assets/ui")
  config.logger = Logger.new($stdout)
end
$stdout.sync = true
specs = [
  { name: "splash", size: [2048, 1152] },
  { name: "button-chrome", size: [768, 256] }
].map do |entry|
  { name: entry[:name], prompt: root.join("assets/ui/prompts/#{entry[:name]}.txt").read,
    image_size: entry[:size], quality: "xhigh", background: "opaque", output_format: "png" }
end
results = Sunburst.generator.generate_batch(specs, on_event: ->(event, **data) {
  puts "#{event}: #{data[:name]}#{data[:job] ? " / #{data[:job].request_id}" : ""}"
})
results.each { |name, result| abort "#{name}: #{result.message}" if result.is_a?(StandardError) }
# macOS's built-in resampler makes the small nine-slice runtime skin from its master.
if RUBY_PLATFORM.include?("darwin")
  abort "Could not resize button chrome" unless system("sips", "-z", "64", "192",
    root.join("assets/ui/button-chrome.png").to_s, "--out",
    root.join("assets/ui/button-chrome-runtime.png").to_s)
else
  puts "Resize button-chrome.png to 192x64 as button-chrome-runtime.png before importing."
end
