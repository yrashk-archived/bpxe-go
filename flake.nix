{
  description = "A flake for building Hello World";

  inputs.nixpkgs.url = github:NixOS/nixpkgs/768a261;

  outputs = { self, nixpkgs }: {

    defaultPackage.x86_64-linux =
      with import nixpkgs { system = "x86_64-linux"; };
      stdenv.mkDerivation {
        name = "bpxe";
        src = self;

        buildInputs = [ go golangci-lint goimports saxon-he ];

        dontInstall = true;

        shellHook = ''
          set -o allexport
          go env > .env
          source .env
          rm .env
          set +o allexport
          export PATH=$PATH:$GOPATH/bin
        '';
      };

  };
}
