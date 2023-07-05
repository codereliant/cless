#!/bin/bash
cd go && ./build.sh
cd ../python && ./build.sh
cd ../nodejs && ./build.sh
cd ../rust && ./build.sh
cd ../java && ./build.sh