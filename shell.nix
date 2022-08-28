{ pkgs ? import ./nix/nixpkgs.nix { }, ... }:

with pkgs;

pkgs.mkShell {
  nativeBuildInputs = [ go_1_19 golangci-lint ];
}
