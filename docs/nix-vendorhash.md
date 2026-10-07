# Updating vendorHash


When `go.mod` or `go.sum` changes, `vendorHash` in `flake.nix` must be updated:

```bash
# 1. Update dependencies
go get github.com/some/pkg@latest
# or: go mod tidy

# 2. Update vendorHash (Nix will compute the new hash)
nix flake update

# 3. Verify everything still works
nix run .#check
```

If `nix flake update` fails with a hash mismatch, set a temporary placeholder and rebuild:

```bash
# In flake.nix, set vendorHash to an empty string temporarily:
vendorHash = "";  # Will show correct hash in error message

# Then run:
nix build .  # Error will show correct hash

# Copy the hash from the error and set it properly:
vendorHash = "sha256-XXXX...";
```

---

_Moved from `AGENTS.md` on 2026-10-07 (carrying-capacity split); AGENTS.md links here._
