{
  description = "nxcd service";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-25.05";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = {self, nixpkgs, flake-utils, ...}:
  flake-utils.lib.eachDefaultSystem (system: {
    packages.default = nixpkgs.legacyPackages.${system}.buildGoModule {
      pname = "nxcd";
      version = "0.0.1";
      src = ./.;
      vendorHash = "sha256-MnvT88VPs7vUlUtK4Qrkci1UsPyo8uqBgbxTMFqDzok=";
    };

    apps.default = {
      type = "app";
      program = "${self.packages.${system}.default}/bin/nxcd";
    };
  }) // {
    nixosModules.default = {config, lib, pkgs, ...}: {
      options.services.nxcd = {
        enable = lib.mkEnableOption "Enable nxcd dns control service";

        private-key-path = lib.mkOption {
          type = lib.types.path;
          default = /etc/ssh/ssh_host_ed25519_key;
          description = "the path of the ssh private key used to access github";
          example = /etc/ssh/ssh_host_ed25519_key;
        };

        poll_duration = lib.mkOption {
          type = lib.types.int;
          default = 60;
          description = "the amount of seconds to wait between each poll";
          example = 60;
        };

        host = lib.mkOption {
          type = lib.types.str;
          default = "";
          description = "the hostname used to deploy the flake configuration";
          example = "srv1";
        };

        repo = lib.mkOption {
          type = lib.types.str;
          default = "";
          description = "the repo name to fetch the flake to deploy from";
          example = "mamaart/srv1";
        };

        branch = lib.mkOption {
          type = lib.types.str;
          default = "main";
          description = "the branch of the repo";
          example = "main";
        };

        matrix = {
          enable = lib.mkEnableOption "Enable matrix";

          homeserver = lib.mkOption {
            type = lib.types.str;
            default = "matrix.org";
            description = "the address of the matrix homeserver";
          };

          username= lib.mkOption {
            type = lib.types.str;
            default = "";
            description = "the username of the matrix client";
          };

          password = lib.mkOption {
            type = lib.types.str;
            default = "";
            description = "the password of the matrix client";
          };

          roomId = lib.mkOption {
            type = lib.types.str;
            default = "";
            description = "the matrix roomId to post the notifications";
          };

        };
      };

      config = lib.mkIf config.services.nxcd.enable {
        systemd.services.nxcd = {
          description = "Exposes nxcd dns control service";
          wantedBy = ["multi-user.target"];
          after = ["network.target"];
          serviceConfig = {
            ExecStart = "${self.packages.${pkgs.system}.default}/bin/nxcd";
            Restart = "always";
            Type = "simple";
            DynamicUser = "yes";
            Environment = [
              "SSH_PRIVATE_KEY_PATH=${toString config.services.nxcd.private-key-path}"
              "REPO=${toString config.services.nxcd.repo}"
              "BRANCH=${toString config.services.nxcd.branch}"
              "HOST=${toString config.services.nxcd.host}"
              "POLL_DURATION=${toString config.services.nxcd.poll_duration}"
            ] ++ lib.mkIf config.services.nxcd.matrix.enable [
              "MATRIX_ENABLED=true"
              "MATRIX_HOMESERVER=${toString config.services.nxcd.matrix.homeserver}"
              "MATRIX_USERNAME=${toString config.services.nxcd.matrix.username}"
              "MATRIX_PASSWORD=${toString config.services.nxcd.matrix.password}"
              "MATRIX_ROOMID=${toString config.services.nxcd.matrix.roomId}"
            ];
          };
        };
      };
    };
  };
}

