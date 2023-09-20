{
src,
pkgs,
gomod2nix,
nixos-lib,
yazbu,
}:

let
  repo = ../.;

  checkgomod2nix = pkgs.writeShellApplication {
    name = "check-gomod2nix";
    runtimeInputs = [ gomod2nix ];
    text = ''
      tmpdir=$(mktemp -d)
      trap 'rm -rf -- "$tmpdir"' EXIT
      gomod2nix --dir "$1" --outdir "$tmpdir"
      if ! diff -q "$tmpdir/gomod2nix.toml" "$1/gomod2nix.toml"; then
        echo '>> gomod2nix.toml is not up to date. Please run:'
        echo '>> $ nix run .#update'
        exit 1
      fi
      echo ">> \"$1/gomod2nix.toml\" is up to date"
    '';
  };

  update = pkgs.writeShellApplication {
    name = "update";
    runtimeInputs = with pkgs; [
      git
      gomod2nix
      helm-docs
    ];
    text = ''
      cd "$(git rev-parse --show-toplevel)"
      gomod2nix
      gomod2nix --dir test/e2e
      helm-docs ./deploy/charts/csi-driver-spiffe
      echo '>> Updated. Please commit the changes.'
    '';
  };

  unit = pkgs.writeShellApplication {
    name = "unit";
    runtimeInputs = with pkgs; [ go ];
    text = ''
      cd "$(git rev-parse --show-toplevel)"
      gofmt -s -l -e .
      go vet -v ./...
      go test --race -v ./config/... ./cmd/... ./internal/...
    '';
  };

  check = pkgs.writeShellApplication {
    name = "check";
    runtimeInputs = with pkgs; [
      git checkgomod2nix unit
    ];
    text = ''
      cd "$(git rev-parse --show-toplevel)"
      check-gomod2nix ${repo}
      check-gomod2nix ${repo}/test/e2e
      unit
    '';
  };

  yazbu-e2e = (pkgs.buildGoApplication {
    name = "yazbu-e2e";
    modules = ../test/e2e/gomod2nix.toml;
    src = pkgs.lib.sourceFilesBySuffices ../test/e2e [ ".go" "go.mod" "go.sum" "gomod2nix.toml" ];
  }).overrideAttrs(old: old // {
    postConfigure = ''
      mv vendor old
      mkdir -p vendor/github.com/joshvanl/yazbu
      cp -r --reflink=auto ${src}/* vendor/github.com/joshvanl/yazbu/.
      cp -r --reflink=auto old/* vendor
    '';
    buildPhase = ''
      go test --race -v -o yazbu-e2e -c
    '';
    installPhase = ''
      mkdir -p $out/bin
      mv yazbu-e2e $out/bin/.
    '';
  });

  mkMinio = port: data: {
    after = [ "network-online.target" ];
    #wantedBy = [ "multi-user.target" ];
    serviceConfig = {
      Type = "simple";
      WorkingDirectory = data;
      ExecStart = "${pkgs.minio}/bin/minio server --json --address :${toString port} ${data}";
      User = "minio";
      Group = "minio";
    };
  };

  e2e = nixos-lib.runTest {
    name = "yazbu end to end tests";
    hostPkgs = pkgs;

    nodes = {
      machine = { pkgs, lib, config, ... }: rec {
        boot.supportedFilesystems = [ "zfs" ];
        networking.hostId = "deadbeef";
        networking.extraHosts = "127.0.0.1 joshvanl-test.localhost";
        virtualisation.emptyDiskImages = [ 4096 ];
        virtualisation.writableStore = true;
        environment.systemPackages = with pkgs; [
          zfs
          parted
          go
          yazbu
          yazbu-e2e
        ];

        users.users.minio = {
          isSystemUser = true;
          group = "minio";
          uid = 2323;
        };
        users.groups.minio.gid = 2323;
        systemd = {
          tmpfiles.rules = [
            "d /var/lib/minio/data1 - minio minio - -"
            "d /var/lib/minio/data2 - minio minio - -"
          ];

          services = {
            minio-1 = mkMinio 9000 "/var/lib/minio/data1";
            minio-2 = mkMinio 9001 "/var/lib/minio/data2";
          };
        };
      };
    };

    testScript = ''
      start_all()

      machine.succeed(
        "modprobe zfs",
        "udevadm settle",
        "parted --script /dev/vdb mklabel msdos",
        "parted --script /dev/vdb -- mkpart primary 1024M -1s",
        "udevadm settle",
        "modprobe zfs",
        "zpool create yazbu /dev/vdb1",
        "zfs create -o mountpoint=legacy yazbu/test-1",
        "zfs create -o mountpoint=legacy yazbu/test-2",
        "mkdir -p /yazbu/testing-1 /yazbu/testing-2",
        "mount -t zfs yazbu/test-1 /yazbu/testing-1",
        "mount -t zfs yazbu/test-2 /yazbu/testing-2",
        "udevadm settle",
      )

      machine.wait_for_open_port(9000)
      machine.wait_until_succeeds("curl -s http://localhost:9000")
      machine.wait_for_open_port(9001)
      machine.wait_until_succeeds("curl -s http://localhost:9001")

      machine.succeed(
        "yazbu-e2e \
          -yazbu-bin=${yazbu}/bin/yazbu \
          -endpoint-1=http://localhost:9000 \
          -endpoint-2=http://localhost:9001 \
          -access-key-1=access-key-1 \
          -access-key-2=access-key-2 \
          -secret-key-1=secret-key-1 \
          -secret-key-2=secret-key-2 \
          -filesystem-1=yazbu/testing-1 \
          -filesystem-2=yazbu/testing-2 \
          -bucketname-1=joshvanl-test-1 \
          -bucketname-2=joshvanl-test-2 \
        "
      )
    '';
  };

  run-e2e = pkgs.writeShellApplication {
    name = "run-e2e.sh";
    runtimeInputs = with pkgs; [ git ];
    text = ''
      cd "$(git rev-parse --show-toplevel)"
      ls -la ${e2e} && exit 0
    '';
  };

in {
  apps = {
    update = {type = "app"; program = "${update}/bin/update";};
    check = {type = "app"; program = "${check}/bin/check";};
    e2e = {type = "app"; program = "${run-e2e}/bin/run-e2e.sh";};
  };
}
