openssl req -x509 -newkey rsa:8102 -nodes \
          -out ./certs/kvcert.pem \
          -keyout ./certs/kvkey.pem -days 365 \
          -subj "/C=BR/ST=Curitiba/L=Parana/O=B\+ Technology/CN=kvstok.com"

