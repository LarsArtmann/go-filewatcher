{
  description = "go-filewatcher - A Go file watching library with debouncing and middleware support";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs = { self, nixpkgs }:
    let
      systems = [ "x86_64-linux" "aarch64-linux" "x86_64-darwin" "aarch64-darwin" ];

      forAllSystems = f: nixpkgs.lib.genAttrs systems (system: f {
        pkgs = nixpkgs.legacyPackages.${system};
        inherit system;
      });
    in
    {
      devShells = forAllSystems ({ pkgs, system }:
        let
          go = pkgs.go_1_24;
        in
        {
          default = pkgs.mkShell {
            name = "go-filewatcher-dev";

            packages = [
              go
              pkgs.gofumpt
              pkgs.gotools
              pkgs.golangci-lint
              pkgs.git
            ];

            shellHook = ''
              echo "go-filewatcher development environment"
              ${go}/bin/go version
              echo ""
              echo "Available commands:"
              echo "  go build ./...                        - Build the project"
              echo "  go test -race ./...                   - Run tests with race detector"
              echo "  go test -v -race ./...                - Run tests with verbose output"
              echo "  go test -race -coverprofile=coverage.out ./..."
              echo "                                        - Run tests with coverage report"
              echo "  golangci-lint run ./...               - Run linter"
              echo "  golangci-lint run --fix ./...         - Run linter with auto-fix"
              echo "  go vet ./...                          - Run go vet"
              echo "  go mod tidy                           - Tidy dependencies"
              echo "  go fmt ./...                          - Format code"
              echo "  go test -bench=. -benchmem ./...      - Run benchmarks"
              echo "  go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out"
              echo "                                        - Generate test coverage report"
              echo "  go test -cover ./...                  - Show test coverage summary"
              echo ""
            '';
          };
        }
      );
    };
}
