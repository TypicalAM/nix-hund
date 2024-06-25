{
  description = "Locate nix development files easily";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let pkgs = import nixpkgs { inherit system; };
      in with pkgs; rec {
        # Development shell
        devShell = mkShell {
          name = "nix-hund";
          nativeBuildInputs = [ go gopls ];
        };

        # Runtime package
        packages.nix-hund = pkgs.buildGoModule rec {
          pname = "nix-hund";
          version = "0.2";

          src = pkgs.fetchFromGitHub
            {
              owner = "TypicalAM";
              repo = "nix-hund";
              rev = "v${version}";
              hash = "sha256-9n/WH6Nodx3uewjtdYWWCW6Kt5uYNttm7NN599nVpmo=";
            } + "/backend";

          vendorHash = "sha256-v4Y6CUxTz59GX7GI8zfI7RsC24/aS0aHW6s6nQzRBkA=";

          meta = {
            description = "Locate nix development files easily";
            homepage = "https://github.com/TypicalAM/nix-hund";
            license = lib.licenses.mit;
            maintainers = with lib.maintainers; [ TypicalAM ];
          };
        };

        # Default package
        defaultPackage = packages.nix-hund;
      }
    );
}
