{ lib
, buildGo119Module
, installShellFiles
}:

buildGo119Module rec {
  pname = "yazbu";
  version = "0.0.1";

  src = ../.;

  vendorSha256 = "sha256-tIr1m7O46eImG/Sa8+tAIzuEl9R79P9HlAC2HX62BQY=";
  subPackages = [ "cmd" ];

  nativeBuildInputs = [ installShellFiles ];

  postInstall = ''
    mv $out/bin/cmd $out/bin/yazbu
    installShellCompletion --cmd yazbu \
      --bash <($out/bin/yazbu completion bash) \
      --fish <($out/bin/yazbu completion fish) \
      --zsh <($out/bin/yazbu completion zsh)
  '';

  meta = with lib; {
    description = "Yet Another ZFS BackerUper.";
    longDescription = ''
      yazbu is a Yet Another ZFS BackerUper.
      Designed to be a simple and easy to use tool for backing up ZFS
      filesystems. Backups decay over time using a configurable decay rate.
    '';
    license = licenses.mit;
  };
}
