{
  description = "go-filewatcher - A Go file watching library with debouncing and middleware support";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts = {
      url = "github:hercules-ci/flake-parts";
      inputs.nixpkgs-lib.follows = "nixpkgs";
    };
    systems.url = "github:nix-systems/default";
    treefmt-nix = {
      url = "github:numtide/treefmt-nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs =
    inputs@{
      self,
      nixpkgs,
      flake-parts,
      systems,
      treefmt-nix,
    }:
    let
      version = self.rev or self.dirtyRev or "dev";
      vendorHash = "sha256-UqxEQDqf0T/SabJZup2FGFtZVbKAaPEMUMcgJB4lt+s=";

      src = nixpkgs.lib.fileset.toSource {
        root = ./.;
        fileset = nixpkgs.lib.fileset.unions [
          ./go.mod
          ./go.sum
          ./doc.go
          ./debouncer.go
          ./debouncer_test.go
          ./errors.go
          ./errors_test.go
          ./event.go
          ./event_test.go
          ./example_test.go
          ./filter.go
          ./filter_gogen.go
          ./filter_gogen_test.go
          ./filter_test.go
          ./fuzz_test.go
          ./metrics.go
          ./metrics_test.go
          ./middleware.go
          ./middleware_test.go
          ./options.go
          ./options_test.go
          ./otel.go
          ./otel_test.go
          ./phantom_types.go
          ./phantom_types_test.go
          ./testing_helpers_test.go
          ./watcher.go
          ./watcher_coverage_test.go
          ./watcher_gitignore.go
          ./watcher_gitignore_test.go
          ./watcher_internal.go
          ./watcher_internal_test.go
          ./watcher_poll.go
          ./watcher_reset_test.go
          ./watcher_selfheal.go
          ./watcher_selfheal_test.go
          ./watcher_test.go
          ./watcher_walk.go
          ./watcher_walk_test.go
          ./benchmark_test.go
          ./backend.go
          ./error_simulation_test.go
          ./fake_backend_test.go
          ./examples
        ];
      };
    in
    flake-parts.lib.mkFlake { inherit inputs; } {
      systems = import systems;

      imports = [
        treefmt-nix.flakeModule
      ];

      perSystem =
        {
          config,
          pkgs,
          lib,
          system,
          ...
        }:
        let
          # Hermetically-built benchstat from golang.org/x/perf (avoids
          # `go run ...@latest` network dependency in bench-diff).
          benchstat = pkgs.buildGoModule {
            pname = "benchstat";
            version = "unstable-2025-07-26";
            src = pkgs.fetchFromGitHub {
              owner = "golang";
              repo = "perf";
              rev = "82a0b07e230d76fa1b3036c383d7a98172f87334";
              hash = "sha256-TOzEoIWofdWlAfKWBS5KWxVpHsn2wx6GZDjACxFZiKI=";
            };
            vendorHash = "sha256-PBvMccuMBBGfJlETw0Xjm5Ojkgg1BS+y9Kc3vwGW5kk=";
            subPackages = [ "cmd/benchstat" ];
            doCheck = false;
          };

          mkApp = name: text: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name text;
                runtimeInputs = with pkgs; [
                  go_1_26
                  golangci-lint
                  gofumpt
                ];
              }
            }/bin/${name}";
          };

          # Like mkApp but also includes benchstat in runtimeInputs.
          mkBenchApp = name: text: {
            type = "app";
            program = "${
              pkgs.writeShellApplication {
                inherit name text;
                runtimeInputs = with pkgs; [
                  go_1_26
                  benchstat
                ];
              }
            }/bin/${name}";
          };
        in
        {
          treefmt = {
            projectRootFile = "go.mod";
            programs = {
              gofumpt.enable = true;
              goimports.enable = true;
              nixfmt.enable = true;
            };
          };

          packages.default = pkgs.buildGoModule {
            pname = "go-filewatcher";
            inherit src version vendorHash;
            doCheck = false;
            meta = {
              description = "High-performance, composable file system watcher for Go";
              homepage = "https://github.com/larsartmann/go-filewatcher";
              license = lib.licenses.mit;
              mainProgram = "go-filewatcher";
              maintainers = [
                {
                  name = "Lars Artmann";
                  github = "LarsArtmann";
                }
              ];
            };
          };

          devShells = {
            default = pkgs.mkShell {
              name = "go-filewatcher";

              packages = [
                pkgs.go_1_26
                pkgs.golangci-lint
                pkgs.gofumpt
                pkgs.golines
                pkgs.gopls
                pkgs.delve
                pkgs.gotools
                pkgs.git
              ];

              shellHook = ''
                echo "go-filewatcher development shell"
                echo "Go version: $(go version)"
                # Redirect Go's temp dir off the tmpfs /tmp onto disk. Long
                # sessions (many nix invocations) previously filled a 24G /tmp;
                # a cache-backed GOTMPDIR avoids that exhaustion.
                export GOTMPDIR="''${XDG_CACHE_HOME:-$HOME/.cache}/go-filewatcher/gotmp"
                mkdir -p "$GOTMPDIR"
                echo "GOTMPDIR=$GOTMPDIR (disk-backed)"
              '';

              GOWORK = "off";
            };

            ci = pkgs.mkShellNoCC {
              packages = [
                pkgs.go_1_26
                pkgs.golangci-lint
              ];

              GOWORK = "off";
            };
          };

          apps = {
            test = mkApp "test" ''
              cd "${self}"
              go test -race -count=1 ./...
            '';

            test-v = mkApp "test-v" ''
              cd "${self}"
              go test -v -race -count=1 ./...
            '';

            lint = mkApp "lint" ''
              cd "${self}"
              golangci-lint run ./...
            '';

            # Explicit test-file linting (--tests is default true, but this app
            # makes the intent clear for CI and local iteration).
            lint-tests = mkApp "lint-tests" ''
              cd "${self}"
              golangci-lint run --tests ./...
            '';

            lint-fix = mkApp "lint-fix" ''
              cd "${self}"
              golangci-lint run --fix ./...
            '';

            vet = mkApp "vet" ''
              cd "${self}"
              go vet ./...
            '';

            # Writes to the working tree, so run in the CALLER's directory
            # (not the read-only nix store copy). Invoke from repo root.
            fmt = mkApp "fmt" ''
              go fmt ./...
              gofumpt -w .
            '';

            bench = mkApp "bench" ''
              cd "${self}"
              go test -bench=. -benchmem -race ./...
            '';

            # Capture a clean benchmark baseline to bench-baseline.txt for later
            # benchstat comparison. Omits -race (it adds overhead and noise) and
            # runs -count=6 so benchstat has enough samples for stable deltas.
            #
            # IMPORTANT: Runs in the CALLER's working directory (not the nix
            # store copy) so the gitignored baseline lands in the repo root.
            # Invoke from the project root: nix run .#bench-baseline
            bench-baseline = mkBenchApp "bench-baseline" ''
              go test -bench=. -benchmem -count=6 -run=^$ ./... | tee bench-baseline.txt
            '';

            # Run fresh benchmarks and diff against bench-baseline.txt with
            # the hermetically-built benchstat (no network needed).
            # Requires a captured baseline (run bench-baseline first).
            #
            # IMPORTANT: Runs in the CALLER's working directory. Invoke from
            # the project root: nix run .#bench-diff
            bench-diff = mkBenchApp "bench-diff" ''
              if [ ! -f bench-baseline.txt ]; then
                echo "bench-baseline.txt not found. Run 'nix run .#bench-baseline' first."
                exit 1
              fi
              go test -bench=. -benchmem -count=6 -run=^$ ./... > "''${TMPDIR:-/tmp}/bench-new.txt"
              benchstat bench-baseline.txt "''${TMPDIR:-/tmp}/bench-new.txt"
            '';

            coverage = mkApp "coverage" ''
              cd "${self}"
              COVERAGE_OUT="''${TMPDIR:-/tmp}/coverage.out"
              go test -coverprofile="$COVERAGE_OUT" ./...
              go tool cover -func="$COVERAGE_OUT"
            '';

            # Writes to the working tree, so run in the CALLER's directory
            # (not the read-only nix store copy). Invoke from repo root.
            tidy = mkApp "tidy" ''
              go mod tidy
            '';

            check = mkApp "check" ''
              cd "${self}"
              echo "Running vet..."
              go vet ./...
              echo "Running lint..."
              golangci-lint run ./...
              echo "Running tests..."
              go test -race -count=1 ./...
              echo "All checks passed."
            '';

            # Writes to the working tree (tidy + fmt), so run in the CALLER's
            # directory — not the read-only nix store copy. Invoke from repo root.
            ci = mkApp "ci" ''
              echo "Running tidy..."
              go mod tidy
              echo "Running fmt..."
              go fmt ./...
              gofumpt -w .
              echo "Running vet..."
              go vet ./...
              echo "Running lint..."
              golangci-lint run ./...
              echo "Running tests..."
              go test -race -count=1 ./...
              echo "CI complete."
            '';

            default = self.apps.${system}.check;
          };

          checks =
            let
              goModules = config.packages.default.goModules;
            in
            {
              format = config.treefmt.build.check self;
              build = config.packages.default;

              test =
                pkgs.runCommand "test"
                  {
                    nativeBuildInputs = [
                      pkgs.go_1_26
                      pkgs.gcc
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
                    nativeBuildInputs = [
                      pkgs.go_1_26
                      pkgs.golangci-lint
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
                    nativeBuildInputs = [
                      pkgs.go_1_26
                      pkgs.gcc
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
                    nativeBuildInputs = [
                      pkgs.go_1_26
                      pkgs.gofumpt
                    ];
                  }
                  ''
                    export GOWORK=off
                    export HOME="$TMPDIR"
                    cp -r "${self}" src && chmod -R u+w src && cd src
                    unformatted=$(gofmt -l .)
                    if [ -n "$unformatted" ]; then
                      echo "Files need formatting:"
                      echo "$unformatted"
                      exit 1
                    fi
                    touch "$out"
                  '';

              examples-build =
                pkgs.runCommand "examples-build"
                  {
                    nativeBuildInputs = [
                      pkgs.go_1_26
                      pkgs.gcc
                    ];
                  }
                  ''
                    export GOWORK=off
                    export HOME="$TMPDIR"
                    cp -r "${self}" src && chmod -R u+w src && cd src
                    ln -s "${goModules}" vendor
                    go build ./examples/...
                    touch "$out"
                  '';
            };
        };

      flake.overlays.default = final: _prev: {
        go-filewatcher = final.callPackage ./package.nix { };
      };
    };
}
