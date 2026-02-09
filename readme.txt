How to setup EKS:

1. Create user with admin permissions (aws -> users -> attach policies -> admins access)

2. Create access keys for admin (admin - security credentials - create)

3. aws configure --profile admin-eks 

4. eksctl create cluster -f create-cluster.yaml --profile admin-eks

5. eksctl create nodegroup -f cloud-public-node-groups.yaml --profile admin-eks

6. Setup completed! EKS created!


daca la kubectl get ... pe cloud da eroare connection refused
tr configura .kube/config, care e setat pe local
aws eks update-kubeconfig --region us-east-1 --name tcc-cluster-cloud --profile admin-eks

# creaza un container temporal (poti sa l fosesti sa testezi lucruri de network)
kubectl run tmp --image=busybox --rm -it -- sh