{
src,
version,
pkgs,
amdPkgs,
armPkgs,
}:

let
  yazbu = sys: (pkgs.buildGoApplication {
    name = "yazbu";
    modules = ../gomod2nix.toml;
    subPackages = [ "cmd" ];
    nativeBuildInputs = [ pkgs.installShellFiles ];
    inherit src version;
    postInstall = ''
      mv $out/bin/cmd $out/bin/yazbu
      find $out -empty -type d -delete
      installShellCompletion --cmd yazbu \
        --bash <($out/bin/yazbu completion bash) \
        --fish <($out/bin/yazbu completion fish) \
        --zsh <($out/bin/yazbu completion zsh)
    '';
    meta = with pkgs.lib; {
      description = "Yet Another ZFS BackerUper.";
      longDescription = ''
        yazbu is a Yet Another ZFS BackerUper.
        Designed to be a simple and easy to use tool for backing up ZFS
        filesystems. Backups decay over time using a configurable decay rate.
      '';
      license = licenses.mit;
    };
  }).overrideAttrs(old: old // { GOARCH = sys; CGO_ENABLED = "0"; });

in {
  inherit yazbu;
  packages = {
    amd64-yazbu = yazbu "amd64";
    arm64-yazbu = yazbu "arm64";
  };
}

