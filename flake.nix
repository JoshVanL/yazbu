{
  description = "yazbu (Yet Another ZFS Backer Upper)";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};

        yazbu = pkgs.callPackage ./nix/build.nix { };

        packageName = "yazbu";
      in {
        packages.${packageName} = yazbu;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [ yazbu ];
        };
    });
}
