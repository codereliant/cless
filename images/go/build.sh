rm -rf docker-gs-ping
git clone git@github.com:docker/docker-gs-ping.git
cd docker-gs-ping && docker build --tag golang-docker -f Dockerfile.multistage .