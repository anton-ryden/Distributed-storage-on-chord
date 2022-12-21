rm *.crt
rm *.key
rm *.csr

# Create CA private key and self-signed certificate
# adding -nodes to not encrypt the private key
openssl req -x509 -newkey rsa:4096 -nodes -days 365 -keyout ca.key -out ca.crt -subj "/C=SE/ST=EU/L=GOTHENBURG/O=DEV/OU=TEST/CN=ca/emailAddress=test@test.com"

echo "CA's self-signed certificate"
openssl x509 -in ca.crt -noout -text