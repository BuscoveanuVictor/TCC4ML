# 1. Creează un director pentru certificate
mkdir -p certs && cd certs

# 2. Generează cheia și certificatul CA (Autoritatea de Certificare)
openssl genrsa -out ca.key 2048
openssl req -x509 -new -nodes -key ca.key -subj "/CN=HybridOrchestratorCA" -days 365 -out ca.crt

# 3. Generează cheia privată pentru serverul tău de Go
openssl genrsa -out tls.key 2048

# 4. Creează un fișier de configurare pentru cererea de semnare (CSR)
# ATENȚIE: 'hybrid-webhook-service' trebuie să fie numele serviciului tău din K8s
cat <<EOF > csr.conf
[req]
req_extensions = v3_req
distinguished_name = req_distinguished_name
[req_distinguished_name]
[ v3_req ]
basicConstraints = CA:FALSE
keyUsage = nonRepudiation, digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names
[alt_names]
DNS.1 = hybrid-webhook-service
DNS.2 = hybrid-webhook-service.default.svc
EOF

# 5. Generează cererea de semnare (CSR)
openssl req -new -key tls.key -out tls.csr -config csr.conf -subj "/CN=hybrid-webhook-service"

# 6. Semnează certificatul serverului cu CA-ul tău
openssl x509 -req -in tls.csr -CA ca.crt -CAkey ca.key \
-CAcreateserial -out tls.crt -days 365 -extensions v3_req -extfile csr.conf