{ pkgs ? import ./nixpkgs.nix { }, ... }:

let
  s3mockFromDockerHub = pkgs.dockerTools.pullImage {
    imageName = "adobe/s3mock";
    imageDigest = "sha256:ef13d8c2e5443dab0db9d7967929fad448f9d35a0a2cf0197a654389f243a69c";
    sha256 = "sha256-gmqzV3KVwAv4EoWQ729OHVdqK8+OI5Spprykfj+S+z4=";
    finalImageTag = "2.4.16";
    finalImageName = "s3mock";
  };

  repo = ../.;

  yazbu = pkgs.callPackage ./build.nix {};
  yazbu-test = pkgs.callPackage ./build-test.nix {};

in pkgs.nixosTest ({
  name = "yazbu end to end tests";

  nodes = {
    machine = { pkgs, lib, ... }: {
      boot = {
        kernelPackages = pkgs.linuxPackages;
        supportedFilesystems = [ "zfs" ];
        zfs.enableUnstable = false;
      };
      networking = {
        hostId = "deadbeef";
        extraHosts = "127.0.0.1 joshvanl-test.localhost";
      };
      virtualisation = {
        emptyDiskImages = [ 4096 ];
        docker.enable = true;
      };
      environment.systemPackages = with pkgs; [
        zfs parted docker go_1_19
        yazbu yazbu-test
      ];
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
      "zpool create yazbu /dev/vdb1",
      "zfs create -o mountpoint=legacy yazbu/test-1",
      "zfs create -o mountpoint=legacy yazbu/test-2",
      "mkdir -p /yazbu/testing-1 /yazbu/testing-2",
      "mount -t zfs yazbu/test-1 /yazbu/testing-1",
      "mount -t zfs yazbu/test-2 /yazbu/testing-2",
      "udevadm settle",
    )

    machine.succeed(
      "docker load --input='${s3mockFromDockerHub}'",
      "docker run -d -p 9090:9090 --hostname s3mock-server-1 --env S3PROXY_AUTHORIZATION=none s3mock:2.4.16",
      "docker run -d -p 9091:9090 --hostname s3mock-server-2 --env S3PROXY_AUTHORIZATION=none s3mock:2.4.16",
    )

    machine.wait_for_open_port(9090)
    machine.wait_for_open_port(9091)
    machine.wait_until_succeeds("curl -s http://localhost:9090")
    machine.wait_until_succeeds("curl -s http://localhost:9091")

    machine.succeed(
      "yazbu-test \
        -yazbu-bin=${yazbu}/bin/yazbu \
        -endpoint-1=http://localhost:9090 \
        -endpoint-2=http://localhost:9091 \
        -filesystem-1=yazbu/testing-1 \
        -filesystem-2=yazbu/testing-2 \
        -bucketname-1=joshvanl-test-1 \
        -bucketname-2=joshvanl-test-2 \
      "
    )
  '';
})
