#!/bin/bash

# loop from 1 to 100
# create string app-${i}.cless.cloud
# put them in an array 
# print the array
for i in {0..100}
do
    s="app-${i}.cless.cloud"
    arr[$i]=$s
done
# join array elements with comma
hosts=$(IFS=, ; echo "${arr[*]}")
echo "127.0.0.1  $hosts"

