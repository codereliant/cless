#!/bin/bash
rm -rf docker-rust-hello
git clone https://github.com/docker/docker-rust-hello
docker build --tag rust-docker -f Dockerfile .