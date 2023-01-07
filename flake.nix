{
  description = "yazbu (Yet Another ZFS Backer Upper)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachSystem [ "x86_64-linux" "aarch64-linux" ]
    (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        yazbu = pkgs.callPackage ./nix/build.nix { };
        e2e = pkgs.callPackage ./nix/test.nix { };
      in rec {
        packages.default = yazbu;
        checks = {
          default = yazbu;
          e2e = e2e;
        };
        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ yazbu ];
        };
    });
}
