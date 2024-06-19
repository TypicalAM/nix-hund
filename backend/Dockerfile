FROM nixos/nix:2.23.0

COPY . /app
WORKDIR /app

RUN nix --extra-experimental-features "nix-command flakes" develop --command go build main.go
