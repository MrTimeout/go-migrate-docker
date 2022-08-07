#!/bin/bash

# Download all the images passed as argument
while read n; do docker image pull $n; done <<< "$(echo java mongo golang | tr ' ' '\n')"