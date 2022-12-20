cd crypto-key
rm *.key

# Generate encryption key
openssl enc -aes-128-cbc -k secret -P -md sha1 > original-sym-private.key

# Output only key into file
cat original-sym-private.key | egrep "key=" | cut -b 5-35 > sym-private.key
