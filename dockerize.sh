#!/bin/bash

echo 'building docker image...'
docker build --file docker/Dockerfile -t keytiles/kt-proxy:1.0.0 -t keytiles/kt-proxy:latest .

echo 'all done'