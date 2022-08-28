{ pkgs }:

pkgs.buildGo119Module rec {
  pname = "yazbu";
  version = "0.0.2";

  src = ../.;

  vendorSha256 = "sha256-dsFA+s6zHyGgl/FqtncwkHfPlJve0gFZP6ooTLN01ZM=";
  subPackages = [ "cmd" ];
  doChecks = false;
  nativeBuildInputs = [ pkgs.installShellFiles ];

  postInstall = ''
    mv $out/bin/cmd $out/bin/yazbu
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
}
