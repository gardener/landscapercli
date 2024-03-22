/*
SPDX-FileCopyrightText: 2021 SAP SE or an SAP affiliate company and Gardener contributors

SPDX-License-Identifier: Apache-2.0
*/
{
  description = "Nix flake for landscaper-cli";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-23.11";
  };

  outputs = {
    self,
    nixpkgs,
    ...
  }: let
    pname = "landscaper-cli";

    # System types to support.
    supportedSystems = ["x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin"];

    # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
    forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

    # Nixpkgs instantiated for supported system types.
    nixpkgsFor = forAllSystems (system: import nixpkgs {inherit system;});
  in {
    # Provide some binary packages for selected system types.
    packages = forAllSystems (system: let
      pkgs = nixpkgsFor.${system};
      inherit (pkgs) stdenv lib;
    in {
      ${pname} = pkgs.buildGo121Module rec {
        inherit pname self;
        version-raw = lib.fileContents ./VERSION;
        version = lib.elemAt (builtins.match "^[v]*([0-9|\.]*)[\-]*(.*)$" version-raw) 0;
        LANDSCAPER_VERSION = lib.elemAt (builtins.match ".*landscaper ([v|0-9|\.]*).*" (lib.fileContents ./go.mod)) 0;

        gitCommit = if (self ? rev) then
          self.rev
        else
          self.dirtyRev;
        state = if (self ? rev) then
          "clean"
        else
          "dirty";

        # This vendorHash represents a dervative of all go.mod dependancies and needs to be adjusted with every change
        # this project includes a vendor folder, so set = null
        vendorHash = null;

        src = ./.;

        ldflags = [
          "-s"
          "-w"
          "-X github.com/gardener/landscapercli/pkg/version.LandscaperCliVersion=${version-raw}"
          "-X github.com/gardener/landscapercli/pkg/version.gitTreeState=${state}"
          "-X github.com/gardener/landscapercli/pkg/version.gitCommit=${gitCommit}"
        ];

        CGO_ENABLED = 0;
        # doCheck = false;
        subPackages = [
          "${pname}"
        ];
        nativeBuildInputs = [pkgs.installShellFiles];

        postInstall = ''
          installShellCompletion --cmd ${pname} \
              --zsh  <($out/bin/${pname} completion zsh) \
              --bash <($out/bin/${pname} completion bash) \
              --fish <($out/bin/${pname} completion fish)
        '';

        meta = with lib; {
          description = "The landscaper-cli interacts with Gardener Landscaper";
          
          longDescription = ''
            The landscape-cli supports users to develop, maintain, and test components processed by the [Gardener Landscaper](https://github.com/gardener/landscaper).
            This comprises the handling of objects like component descriptors, blueprints, installations, etc.
          '';
          homepage = "https://github.com/gardener/landscapercli";
          license = licenses.asl20;
          platforms = supportedSystems;
        };
      };
    });

    # Add dependencies that are only needed for development
    devShells = forAllSystems (system: let
      pkgs = nixpkgsFor.${system};
    in {
      default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_21   # golang 1.21
          gopls     # go language server
          gotools   # go imports
          go-tools  # static checks
          gnumake   # standard make
        ];
      };
    });

    # The default package for 'nix build'
    defaultPackage = forAllSystems (system: self.packages.${system}.${pname});
  };
}
