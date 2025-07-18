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
      vendorHash = "sha256-OR5+te37m1/y4Q2ot169WubRJdW6b0SkMxwmQ93eeDE=";
    };

    apps.default = {
      type = "app";
      program = "${self.packages.${system}.default}/bin/nxcd";
    };
  }) // {
    nixosModules.default = {config, lib, pkgs, ...}: {
      options.services.nxcd = {
        enable = lib.mkEnableOption "Enable nxcd dns control service";

        configFile = lib.mkOption {
          type = lib.types.path;
          description = "the path of the config (yaml) file";
        };

        private-key-raw = lib.mkOption {
          type = lib.types.path;
          default = /etc/ssh/ssh_host_ed25519_key;
          description = "the path of the ssh private key used to access github";
          example = /etc/ssh/ssh_host_ed25519_key;
        };

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
          description = "Exposes nxcd nix deployment service";
          wantedBy = ["multi-user.target"];
          after = ["network.target"];
          serviceConfig = {
            ExecStart = "${self.packages.${pkgs.system}.default}/bin/nxcd";
            Restart = "always";
            Type = "simple";
            DynamicUser = "yes";
            LoadCredential = 
              lib.optional (config.services.nxcd.private-key-raw != null) "ssh_key:${toString config.services.nxcd.private-key-raw}"
              ++ 
              lib.optional (config.services.nxcd.configFile != null) "config:${toString config.services.nxcd.configFile}";
            Environment = 
              if (config.services.nxcd.configFile != null) then 
                [ "APP_CONFIG=/run/credentials/%N.service/config" ]
              else
              [
                "GIT_SSH_PRIVATE_KEY_PATH=/run/credentials/%N.service/ssh_key"
                "GIT_REPO=${toString config.services.nxcd.repo}"
                "GIT_BRANCH=${toString config.services.nxcd.branch}"
                "HOST=${toString config.services.nxcd.host}"
                "POLL_DURATION=${toString config.services.nxcd.poll_duration}"
                "MATRIX_ENABLED=${toString config.services.nxcd.matrix.enable}"
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

