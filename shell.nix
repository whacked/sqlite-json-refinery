{ pkgs ? import <nixpkgs> {} }:

let

  ws4sqlite = pkgs.stdenv.mkDerivation {
    pname = "ws4sqlite";
    version = "0.16.2";
    
    src = pkgs.fetchurl {
      url = "https://github.com/proofrock/ws4sqlite/releases/download/v0.16.2/ws4sqlite-v0.16.2-linux-amd64.tar.gz";
      hash = "sha256-2scK/60xHGM5VkQq0ywl0eoEpL/yarebNQbkcYBmxyI=";
    };
    
    sourceRoot = ".";
    
    installPhase = ''
      mkdir -p $out/bin
      cp ws4sqlite $out/bin/
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

    ws4sqlite  # Added here
  ] ++ [
    # frontend
    pkgs.nodePackages.pnpm

  ];

  shellHook = ''
    export PATH=$PWD/result/bin:$PATH
  '';
}
