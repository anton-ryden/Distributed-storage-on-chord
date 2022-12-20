cd certs
rm server.*
rm client.*


# Create Web Server private key and CSR
# adding -nodes to not encrypt the private key
openssl req -newkey rsa:4096 -nodes -keyout server.key -out server.csr -subj "/C=SE/ST=EU/L=GOTHENBURG/O=DEV/OU=SERVER/CN=server/emailAddress=test2@test.com"

# Sign the Web Server Certificate Request (CSR)
openssl x509 -req -extfile <(printf "subjectAltName=DNS:localhost,IP:$1") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt

echo "Server's signed certificate"
openssl x509 -in server.crt -noout -text

# Verify certificate
echo "Verifying certificate"
openssl verify -CAfile ca.crt server.crt

# Generate client's private key and certificate signing request (CSR)
openssl req -newkey rsa:4096 -nodes -keyout client.key -out client.csr -subj "/C=SE/ST=EUROPE/L=GOTHENBURG/O=DEV/OU=CLIENT/CN=client/emailAddress=someclient@gmail.com"

#  Sign the Client Certificate Request (CSR)
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out client.crt

echo "Client's signed certificate"
openssl x509 -in client.crt -noout -text