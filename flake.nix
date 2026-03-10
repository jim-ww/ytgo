{
  description = "ytgo nix flake";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-parts.url = "github:hercules-ci/flake-parts";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = inputs @ {
    nixpkgs,
    flake-parts,
    flake-utils,
    ...
  }:
    flake-parts.lib.mkFlake {inherit inputs;} {
      systems = flake-utils.lib.defaultSystems;

      perSystem = {pkgs, ...}: let
        runtimeDeps = [pkgs.mpv];
        ytgoPkg = pkgs.buildGoModule {
          pname = "ytgo";
          version = "1.0";
          src = pkgs.lib.cleanSource ./.;
          vendorHash = "sha256-xGRNbgRzV1dT4q3QtOTr3ybAmxI5OPYiH8PHdCoZBUQ=";

          buildInputs = runtimeDeps;
          nativeBuildInputs = [pkgs.go pkgs.makeWrapper] ++ runtimeDeps;
          nativeCheckInputs = [pkgs.go pkgs.makeWrapper] ++ runtimeDeps;

          buildPhase = ''
            go build -o ytgo main.go
          '';

          installPhase = ''
            install -Dm755 ytgo $out/bin/ytgo
            wrapProgram $out/bin/ytgo --prefix PATH : ${pkgs.lib.makeBinPath runtimeDeps}
          '';
        };
      in {
        packages.default = ytgoPkg;

        devShells.default = pkgs.mkShell {
          buildInputs = [pkgs.go] ++ runtimeDeps;
        };

        apps.default = flake-utils.lib.mkApp {
          drv = ytgoPkg;
        };
      };
    };
}
