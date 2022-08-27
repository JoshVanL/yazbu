{ pkgs ? import ./nixpkgs.nix { }, ... }:

let
  s3proxyFromDockerHub = pkgs.dockerTools.pullImage {
    imageName = "andrewgaul/s3proxy";
    imageDigest = "sha256:f5d4fc2ccad9c6f90b226d2d47d75311dbdeb40793bb094cc0bfb30235a9c250";
    sha256 = "sha256-dIlEVi4/RJjCSzHYHN9RO1CvaEu/3j7WMpwJKWSPr+o=";
    finalImageTag = "sha-ba0fd6d";
    finalImageName = "s3proxy";
  };

  repo = ../.;

  yazbu = pkgs.callPackage ./build.nix {};
  yazbu-test = pkgs.callPackage ./build-test.nix {};

in pkgs.nixosTest ({
  name = "yazbu integration tests";

  nodes = {
    machine = { pkgs, lib, ... }: {
      boot = {
        kernelPackages = pkgs.linuxPackages;
        supportedFilesystems = [ "zfs" ];
        zfs.enableUnstable = false;
      };
      networking = {
        hostId = "deadbeef";
        extraHosts = "localhost joshvanl-test.localhost";
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
      "docker load --input='${s3proxyFromDockerHub}'",
      "docker run -d -p 80:80 --hostname s3proxy-server-1 --env S3PROXY_AUTHORIZATION=none s3proxy:sha-ba0fd6d",
      "docker run -d -p 81:80 --hostname s3proxy-server-2 --env S3PROXY_AUTHORIZATION=none s3proxy:sha-ba0fd6d",
    )

    machine.wait_for_open_port(80)
    machine.wait_for_open_port(81)
    machine.wait_until_succeeds("curl -s http://localhost:80")
    machine.wait_until_succeeds("curl -s http://localhost:81")

    machine.succeed(
      "yazbu-test \
        -endpoint-1=http://localhost:80 \
        -endpoint-2=http://localhost:81 \
        -filesystem-1=yazbu/testing-1 \
        -filesystem-2=yazbu/testing-2 \
        -bucketname-1=joshvanl-test-1 \
        -bucketname-2=joshvanl-test-2 \
      "
    )
  '';
})
