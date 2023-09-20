{
  description = "yazbu (Yet Another ZFS Backer Upper)";

  inputs = {
    nixpkgs.url = "nixpkgs/nixos-unstable";
    utils.url = "github:numtide/flake-utils";

    gomod2nix = {
      url = "github:tweag/gomod2nix";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.utils.follows = "utils";
    };
  };

  outputs = { self, nixpkgs, utils, gomod2nix }:
  let
    lib = nixpkgs.lib;
    targetSystems = with utils.lib.system; [
      x86_64-linux
      x86_64-darwin
      aarch64-linux
      aarch64-darwin
    ];

    src = nixpkgs.lib.sourceFilesBySuffices ./. [ ".go" "go.mod" "go.sum" "gomod2nix.toml" ];
    version = "0.0.2";

  in utils.lib.eachSystem targetSystems (system:
    let
      overlays = lib.mapAttrsToList (name: _: import ./nix/overlays/${name})
      (lib.filterAttrs
        (name: entryType: lib.hasSuffix ".nix" name) (builtins.readDir ./nix/overlays)
      ) ++ [ gomod2nix.overlays.default ];

      pkgs = import nixpkgs { inherit system overlays; };
      nixos-lib = import (nixpkgs + "/nixos/lib") { };
      amdPkgs = import nixpkgs { inherit overlays; system = "x86_64-linux"; };
      armPkgs = import nixpkgs { inherit overlays; system = "aarch64-linux"; };

      localSystem = if pkgs.stdenv.hostPlatform.isAarch64 then "arm64" else "amd64";

      build = import ./nix/build.nix { inherit src version pkgs amdPkgs armPkgs; };

      ci = import ./nix/ci.nix {
        inherit src pkgs nixos-lib;
        gomod2nix = (gomod2nix.packages.${system}.default);
        yazbu = build.yazbu localSystem;
      };

    in {
      packages = {
        default = (build.yazbu localSystem);
        yazbu = (build.yazbu localSystem);
      };

      apps = {
        default = {type = "app"; program = "${self.packages.${system}.default}/bin/yazbu"; };
      } // ci.apps;

      devShells.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go
          gopls
          gotools
          go-tools
          gomod2nix.packages.${system}.default
        ];
      };
  });
}

