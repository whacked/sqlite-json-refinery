{ pkgs ? import <nixpkgs> {} }:

let

  nix_shortcuts = import (pkgs.fetchurl {
    url = "https://raw.githubusercontent.com/whacked/setup/f6338a7796e24bdca23a3e5a2297309dffe84322/bash/nix_shortcuts.nix.sh";
    hash = "sha256-jLbvJ52h12eug/5Odo04kvHqwOQRzpB9X3bUEB/vzxc=";
  }) { inherit pkgs; };

  ws4sqlite = pkgs.stdenv.mkDerivation {
    pname = "ws4sqlite";
    version = "0.16.3";
    
    src = if pkgs.stdenv.isDarwin then
      pkgs.fetchzip {
        url = "https://github.com/proofrock/ws4sqlite/releases/download/v0.16.3/ws4sqlite-v0.16.3-darwin-arm64.zip";
        hash = "sha256-MLlDrMi0zum4sH7Eks1xJ94PHyqmazaFIA2L2olgjWY=";
      }
    else
      pkgs.fetchzip {
        url = "https://github.com/proofrock/ws4sqlite/releases/download/v0.16.3/ws4sqlite-v0.16.3-linux-amd64.tar.gz";
        hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";
    };
    
    sourceRoot = ".";
    
    installPhase = ''
      mkdir -p $out/bin
      cp ./source/ws4sqlite $out/bin/
      chmod +x $out/bin/ws4sqlite
    '';
  };

  go-jsonschema = pkgs.stdenv.mkDerivation {
    pname = "go-jsonschema";
    version = "0.16.0";

    src = pkgs.fetchurl (
      if pkgs.stdenv.isDarwin then {
        url = "https://github.com/omissis/go-jsonschema/releases/download/v0.16.0/go-jsonschema_Darwin_x86_64.tar.gz";
        hash = "sha256-nbxSUGp3jLU3pD141P6pwPr8l9HR4QwvNTz2oTdbNX4=";
      } else {
        url = "https://github.com/omissis/go-jsonschema/releases/download/v0.16.0/go-jsonschema_Linux_x86_64.tar.gz";
        hash = "sha256-+gLwQURYBwEGIHKXeSIqF9GRRnd0iyt6d7KqlQQ3HQw=";
      }
    );

    nativeBuildInputs = [ pkgs.gnutar ];

    sourceRoot = ".";

    installPhase = ''
      mkdir -p $out/bin
      cp go-jsonschema $out/bin/
      chmod +x $out/bin/go-jsonschema
    '';

    meta = with pkgs.lib; {
      description = "A tool to generate Go types from JSON Schema";
      homepage = "https://github.com/omissis/go-jsonschema";
      license = licenses.mit;
      platforms = platforms.linux ++ platforms.darwin;
    };
  };
in pkgs.mkShell {
  packages = [
    pkgs.go
    pkgs.sqlite
    go-jsonschema
    pkgs.oapi-codegen

    ws4sqlite
    pkgs.static-web-server
  ] ++ [
    # frontend
    pkgs.nodePackages.nodejs
    pkgs.nodePackages.pnpm

  ] ++ nix_shortcuts.buildInputs;

  shellHook = nix_shortcuts.shellHook + ''
    export PATH=$PWD/result/bin:$PATH

    generate-ws4sqlite-conf() {  # generate yaml file for ws4sqlite
      if [ $# -ne 1 ]; then
        echo "need path to database"
        return
      fi
      database_file=$1
      yaml_file=''${database_file%.*}.yaml
      echo "generating file at $yaml_file"
      echo "corsOrigin: '*'" >> $yaml_file
      echo "readOnly: true" >> $yaml_file
    }

    alias serve-database="ws4sqlite -db"
    alias serve-static-media="static-web-server --root $STATICDATA/mmc/DCIM --port 7002"
  '' + ''
    echo-shortcuts ${__curPos.file}
  '';
}
