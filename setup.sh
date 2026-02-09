cd 
cd TCC

cd certs

chmod +x setup_secrets.sh
./setup_secrets.sh

cd ..

cd webhook
chmod +x run_webhook.sh
./run_webhook.sh

cd ..

cd dispacher
chmod +x run_dispatcher.sh
./run_dispatcher.sh