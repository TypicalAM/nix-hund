FROM nixos/nix:2.23.0
LABEL maintainer="TypicalAM"

COPY . /src
WORKDIR /src

RUN nix --extra-experimental-features "nix-command flakes" build \
	&& cp /src/result/bin/nix-hund /bin/nix-hund

RUN mkdir /hund \
	&& chown -R nobody:nobody /hund

USER nobody
EXPOSE 1323
ENTRYPOINT [ "/bin/nix-hund" ]
CMD        [ "-cache_dir=/hund" ]
