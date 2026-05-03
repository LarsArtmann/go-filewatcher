{
  description = "go-filewatcher - A Go file watching library with debouncing and middleware support";

  # This flake provides a reproducible development environment.
  # Go 1.26.1 and golangci-lint are provided by the system nixpkgs.

  inputs.nixpkgs.url = "nixpkgs";

  outputs = { self, nixpkgs }:
    let
      supportedSystems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      forEachSystem = f: nixpkgs.lib.genAttrs supportedSystems (system: f {
        inherit system;
        pkgs = nixpkgs.legacyPackages.${system};
      });
    in
    {
      devShells = forEachSystem ({ pkgs, system }: {
        default = pkgs.mkShell {
          name = "go-filewatcher";

          packages = with pkgs; [
            go
            golangci-lint
            gofumpt
            git
          ];

          shellHook = ''
            echo "go-filewatcher development shell"
            echo "Go version: $(go version)"
            echo "golangci-lint version: $(golangci-lint --version)"
            echo ""
            echo "Available commands (GOWORK=off is set automatically):"
            echo "  build              - go build ./..."
            echo "  test               - go test -race ./..."
            echo "  test-v             - go test -v -race ./..."
            echo "  test-cover         - Generate coverage report"
            echo "  lint               - golangci-lint run ./..."
            echo "  lint-fix           - golangci-lint run --fix ./..."
            echo "  vet                - go vet ./..."
            echo "  tidy               - go mod tidy"
            echo "  fmt                - go fmt ./..."
            echo "  bench              - go test -bench=. -benchmem ./..."
            echo "  coverage           - go test -coverprofile=coverage.out ./..."
            echo "  coverage-summary   - go test -cover ./..."
            echo "  check              - vet + lint + test"
            echo "  ci                 - tidy + fmt + vet + lint + test"
          '';

          GOWORK = "off";
        };
      });
    };
}
