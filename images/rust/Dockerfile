FROM rust:1.70.0 as builder
WORKDIR /usr/src/docker-rust-hello
COPY docker-rust-hello .
RUN cargo install --path .
FROM ubuntu:23.10
COPY --from=builder /usr/local/cargo/bin/docker-rust-hello /usr/local/bin/docker-rust-hello
ENV ROCKET_ADDRESS=0.0.0.0
ENV ROCKET_PORT=8080
EXPOSE 8080
CMD ["docker-rust-hello"]