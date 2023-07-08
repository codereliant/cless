#!/bin/bash
cd images/go && ./build.sh
cd ../python && ./build.sh
cd ../nodejs && ./build.sh
cd ../rust && ./build.sh
cd ../java && ./build.sh