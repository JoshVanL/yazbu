{ pkgs }:

with pkgs;

let
  src = ../.;

  go-modules = stdenv.mkDerivation (let modArgs = {
    name = "yazbu-test-go-modules";

    nativeBuildInputs = [ go_1_19 git cacert ];

    inherit src;
    inherit (go_1_19) GOOS GOARCH;

    GO111MODULE = "on";
    configurePhase = ''
      export GOCACHE=$TMPDIR/go-cache
      export GOPATH="$TMPDIR/go"
    '';

    buildPhase = "go mod vendor";
    installPhase = "cp -r --reflink=auto vendor $out";
    dontFixup = true;
  }; in modArgs // (
      {
        outputHashMode = "recursive";
        outputHashAlgo = "sha256";
        outputHash = "sha256-KDMFbHo341pIvyoi3i/l46In+gCAhyIwwkDWl2yPSzo=";
      }
  ) // modArgs);

in stdenv.mkDerivation {
  name = "yazbu-test";
  src = lib.sourceFilesBySuffices ./.. [ ".go" ".mod" ".sum" ];

  nativeBuildInputs = [ go ];

  configurePhase = ''
    export GOCACHE=$TMPDIR/go-cache
    export GOPATH="$TMPDIR/go"
    export GOPROXY=off
    export GOSUMDB=off
    rm -rf vendor
    cp -r --reflink=auto ${go-modules} vendor
    export GOPROXY=file://${go-modules}
  '';

  buildPhase = ''
    gofmt -s -l -e .
    go vet -v ./...
    go test --race -v ./cmd/... ./internal/...
    go test --race -v -o yazbu-test -c ./test/e2e/.
  '';
  installPhase = ''
    mkdir -p $out/bin
    mv yazbu-test $out/bin/.
  '';
}
