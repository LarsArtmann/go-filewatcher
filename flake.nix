{
  description = "go-filewatcher - A Go file watching library with debouncing and middleware support";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";

  outputs =
    { self, nixpkgs }:
    let
      supportedSystems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forEachSystem =
        f:
        nixpkgs.lib.genAttrs supportedSystems (
          system:
          f {
            inherit system;
            pkgs = nixpkgs.legacyPackages.${system};
          }
        );
    in
    {
      packages = forEachSystem (
        { pkgs, system }:
        {
          default = pkgs.buildGoModule {
            pname = "go-filewatcher";
            version = "0.0.0";
            src = self;
            vendorHash = "sha256-k7cUWDloqRcDOL2Npmh3+9NOhiV5DVELIH5PuaGFrDs=";
            doCheck = false;
            meta = with pkgs.lib; {
              description = "High-performance, composable file system watcher for Go";
              homepage = "https://github.com/larsartmann/go-filewatcher";
              license = licenses.mit;
            };
          };
        }
      );

      devShells = forEachSystem (
        { pkgs, system }:
        {
          default = pkgs.mkShell {
            name = "go-filewatcher";

            packages = with pkgs; [
              go_1_26
              golangci-lint
              gofumpt
              golines
              gopls
              delve
              gotools
              git
            ];

            shellHook = ''
              alias check='nix run .#check'
              alias ci='nix run .#ci'
              alias lint='nix run .#lint'
              alias lint-fix='nix run .#lint-fix'
              alias test='nix run .#test'

              echo "go-filewatcher development shell"
              echo "Go version: $(go version)"
              echo "golangci-lint version: $(golangci-lint --version)"
              echo ""
              echo "Commands (nix run .#<name> or alias):"
              echo "  check       - vet + lint + test"
              echo "  ci          - tidy + fmt + vet + lint + test"
              echo "  lint-fix    - Auto-fix linter issues"
              echo "  test        - Run tests with -race"
              echo "  test-v      - Run tests with -race -v"
              echo "  lint        - Run linter"
              echo "  bench       - Run benchmarks"
              echo "  coverage    - Generate coverage report"
              echo "  fmt         - Format Go code"
              echo "  tidy        - Run go mod tidy"
            '';

            GOWORK = "off";
          };
        }
      );

      apps = forEachSystem (
        { pkgs, system }:
        {
          test = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "test" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go test -race -count=1 ./...
            ''}/bin/test";
          };

          test-v = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "test-v" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go test -v -race -count=1 ./...
            ''}/bin/test-v";
          };

          lint = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "lint" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.golangci-lint}/bin/golangci-lint run ./...
            ''}/bin/lint";
          };

          lint-fix = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "lint-fix" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.golangci-lint}/bin/golangci-lint run --fix ./...
            ''}/bin/lint-fix";
          };

          vet = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "vet" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go vet ./...
            ''}/bin/vet";
          };

          fmt = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "fmt" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go fmt ./...
              ${pkgs.gofumpt}/bin/gofumpt -w .
            ''}/bin/fmt";
          };

          bench = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "bench" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go test -bench=. -benchmem ./...
            ''}/bin/bench";
          };

          coverage = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "coverage" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go test -coverprofile=coverage.out ./...
              ${pkgs.go_1_26}/bin/go tool cover -func=coverage.out
            ''}/bin/coverage";
          };

          tidy = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "tidy" ''
              cd "${self}"
              export GOWORK=off
              ${pkgs.go_1_26}/bin/go mod tidy
            ''}/bin/tidy";
          };

          check = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "check" ''
              cd "${self}"
              export GOWORK=off
              echo "Running vet..."
              ${pkgs.go_1_26}/bin/go vet ./...
              echo "Running lint..."
              ${pkgs.golangci-lint}/bin/golangci-lint run ./...
              echo "Running tests..."
              ${pkgs.go_1_26}/bin/go test -race -count=1 ./...
              echo "All checks passed."
            ''}/bin/check";
          };

          ci = {
            type = "app";
            program = "${pkgs.writeShellScriptBin "ci" ''
              cd "${self}"
              export GOWORK=off
              echo "Running tidy..."
              ${pkgs.go_1_26}/bin/go mod tidy
              echo "Running fmt..."
              ${pkgs.go_1_26}/bin/go fmt ./...
              echo "Running vet..."
              ${pkgs.go_1_26}/bin/go vet ./...
              echo "Running lint..."
              ${pkgs.golangci-lint}/bin/golangci-lint run ./...
              echo "Running tests..."
              ${pkgs.go_1_26}/bin/go test -race -count=1 ./...
              echo "CI complete."
            ''}/bin/ci";
          };

          default = self.apps.${system}.check;
        }
      );

      checks = forEachSystem (
        { pkgs, system }:
        let
          goModules = self.packages.${system}.default.goModules;
        in
        {
          build = self.packages.${system}.default;

          test =
            pkgs.runCommand "test"
              {
                nativeBuildInputs = with pkgs; [
                  go_1_26
                  gcc
                ];
              }
              ''
                export GOWORK=off
                export HOME="$TMPDIR"
                cp -r "${self}" src && chmod -R u+w src && cd src
                ln -s "${goModules}" vendor
                go test -race -count=1 ./...
                touch "$out"
              '';

          lint =
            pkgs.runCommand "lint"
              {
                nativeBuildInputs = with pkgs; [
                  go_1_26
                  golangci-lint
                ];
              }
              ''
                export GOWORK=off
                export HOME="$TMPDIR"
                cp -r "${self}" src && chmod -R u+w src && cd src
                ln -s "${goModules}" vendor
                golangci-lint run ./...
                touch "$out"
              '';

          vet =
            pkgs.runCommand "vet"
              {
                nativeBuildInputs = with pkgs; [
                  go_1_26
                  gcc
                ];
              }
              ''
                export GOWORK=off
                export HOME="$TMPDIR"
                cp -r "${self}" src && chmod -R u+w src && cd src
                ln -s "${goModules}" vendor
                go vet ./...
                touch "$out"
              '';

          go-fmt =
            pkgs.runCommand "go-fmt"
              {
                nativeBuildInputs = with pkgs; [
                  go_1_26
                  gofumpt
                ];
              }
              ''
                export GOWORK=off
                export HOME="$TMPDIR"
                cd "${self}"
                unformatted=$(gofmt -l .)
                if [ -n "$unformatted" ]; then
                  echo "Files need formatting:"
                  echo "$unformatted"
                  exit 1
                fi
                touch "$out"
              '';
        }
      );

      formatter = forEachSystem ({ pkgs, ... }: pkgs.nixfmt);
    };
}
