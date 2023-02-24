openssl req -x509 -newkey rsa:4096 -nodes \
          -out ./certs/kvcert.pem \
          -keyout ./certs/kvkey.pem -days 365
