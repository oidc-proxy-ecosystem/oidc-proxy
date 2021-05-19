#!/bin/env bash
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 -keyout ssl/server.key -out ssl/server.crt -subj "/C=JP/ST=Nara/L=Kashiba/O=n-creativesystem/CN=proxy.n-creativesystem.com"
