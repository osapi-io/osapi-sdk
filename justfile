# Optional modules: mod? allows `just fetch` to work before .just/remote/ exists.
mod? go '.just/remote/go.mod.just'

# --- Fetch ---

# Fetch shared justfiles from osapi-io-justfiles
fetch:
    mkdir -p .just/remote
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/go.mod.just -o .just/remote/go.mod.just
    curl -sSfL https://raw.githubusercontent.com/osapi-io/osapi-io-justfiles/refs/heads/main/go.just -o .just/remote/go.just

# --- Top-level orchestration ---

# Install all dependencies
deps:
    just go::deps
    just go::mod

# Run all tests
test:
    just go::test

# Generate code
generate:
    go tool github.com/retr0h/gilt/v2 overlay
    redocly join --prefix-tags-with-info-prop title -o pkg/osapi/gen/api.yaml pkg/osapi/gen/*/gen/api.yaml
    just go::generate
