#!/bin/sh

appctl run --app-id nodeapp \
    --app-protocol http \
    --app-port 3000 \
    --app-http-port 3500 \
    --components-path ./config \
    --log-level debug \
    node app.js