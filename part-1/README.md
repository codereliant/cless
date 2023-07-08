# cLess

# in this part we will just try to get up and running a cless service.
- be able to start docker container on request
- proxy requests to the right container
- hardcoding host to container name


# build images locally so they are available for docker
```bash
cd .. && ./build_images.sh
```

## Setup 
add this to /etc/hosts
```
127.0.0.1       golang.cless.cloud
127.0.0.1       python.cless.cloud
127.0.0.1       java.cless.cloud
127.0.0.1       nodejs.cless.cloud
127.0.0.1       rust.cless.cloud
```

## Run Server
```bash
go run build && ./cless
```

## Test or Browse
```bash
curl http://rust.cless.cloud/
curl http://java.cless.cloud/
```

## architecture
![Diagram](diagram.jpg)
