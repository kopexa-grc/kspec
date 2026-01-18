schema_version = 1

project {
  license        = "Elastic-2.0"
  copyright_holder = "Kopexa GmbH"
  copyright_year = 2024

  # Patterns to ignore when adding headers
  header_ignore = [
    # Vendor and external dependencies
    "vendor/**",

    # Generated files
    "**/*_gen.go",
    "**/*.pb.go",
    "**/mock_*.go",
    "**/mocks/**",

    # Test fixtures and data
    "**/*.json",
    "**/*.yaml",
    "**/*.yml",
    "**/*.md",
    "**/*.txt",
    "**/*.html",
    "**/*.css",

    # Build artifacts
    "dist/**",
    "bin/**",
    "build/**",

    # IDE and tool configs
    ".vscode/**",
    ".idea/**",
    ".github/**",

    # Go module files
    "go.mod",
    "go.sum",

    # Other config files
    ".goreleaser.yml",
    ".golangci.yml",
    "Makefile",
    ".gitignore",
    ".copywrite.hcl",
  ]
}
