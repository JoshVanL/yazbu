let
  nixpkgs = fetchTarball "https://github.com/NixOS/nixpkgs/archive/376485f1ef53272bb0474d6554ae0a5bb116533b.tar.gz";
  pkgs = import nixpkgs {};

  s3mockFromDockerHub = pkgs.dockerTools.pullImage {
    imageName = "adobe/s3mock";
    imageDigest = "sha256:151bce411c68caa7055097fe9145631e8caad8ad57596d62391539dc116277cb";
    sha256 = "sha256-gmqzV3KVwAv4EoWQ729OHVdqK8+OI5Spprykfj+S+z4=";
    finalImageTag = "2.4.16";
    finalImageName = "s3mock";
  };

  testScript = pkgs.writeShellScriptBin "test.sh" (builtins.readFile ./test.sh);

  yazbu = pkgs.callPackage ./build.nix {};

  testConfigFile = ''
    buckets:
    - name: joshvanl-test
      region: auto
      storageClass: STANDARD
      endpoint: http://localhost:9090
      accessKey: server-1-access-key
      secretKey: server-1-secret-key
    - name: joshvanl-test
      region: auto
      storageClass: STANDARD
      endpoint: http://localhost:9091
      accessKey: server-2-access-key
      secretKey: server-2-secret-key
    filesystems:
    - yazbu/testing-1
    - yazbu/testing-2
    cadence:
      incrementalPerLastFull: 7
      fullLast45Days: 10
      full45To182Days: 10
      full182To365Days: 5
      fullPer365Over365Days: 4
  '';

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
      environment = {
        etc = { "yazbu/config.yaml" = { text = testConfigFile; }; };
        systemPackages = with pkgs; [
          zfs parted docker
          testScript yazbu
        ];
      };
    };
  };

  testScript = ''
    start_all()

    machine.succeed(
      "modprobe zfs",
      "zpool status",
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
      "zfs list",
    )

    machine.succeed(
      "docker load --input='${s3mockFromDockerHub}'",
      "docker run -d -p 9090:9090 --hostname s3mock-server-1 s3mock:2.4.16",
      "docker run -d -p 9091:9090 --hostname s3mock-server-2 s3mock:2.4.16",
    )

    machine.wait_for_open_port(9090)
    machine.wait_for_open_port(9091)
    machine.wait_until_succeeds("curl http://localhost:9090")
    machine.wait_until_succeeds("curl http://localhost:9091")

    machine.succeed(
      "YAZBU_E2E=1 test.sh"
    )
  '';
})
