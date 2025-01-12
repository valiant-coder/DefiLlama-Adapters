#!/bin/bash

mkdir -p certs

cat > certs/openssl.cnf << EOF
[req]
default_bits = 4096
prompt = no
default_md = sha256
req_extensions = req_ext
distinguished_name = dn
[dn]
C = US
ST = California
L = San Francisco
O = ExApp
OU = Dev
CN = localhost
[req_ext]
subjectAltName = @alt_names
[alt_names]
DNS.1 = localhost
DNS.2 = *.localhost
IP.1 = 127.0.0.1
IP.2 = 192.168.43.141
EOF

openssl req -x509 \
    -newkey rsa:4096 \
    -keyout certs/server.key \
    -out certs/server.crt \
    -days 365 \
    -nodes \
    -config certs/openssl.cnf \
    -extensions req_ext

chmod 600 certs/server.key
chmod 644 certs/server.crt


echo "key: certs/server.key"
echo "crt: certs/server.crt" 
