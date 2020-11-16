# can-i-pull
`canipull` is a simple kubectl plugin to validate if your Kubernetes cluster running on Azure (AKS or self-hosted) has correct setup to pull container images from Azure Container Registry.

## Install
On linux / mac, run:

```bash
wget https://raw.githubusercontent.com/yangl900/canipull/main/plugin/kubectl-check_acr && \ 
chmod +x ./kubectl-check_acr && \
sudo mv -f ./kubectl-check_acr /usr/local/bin
```

## Usage
Following command will perform the authorization check as well as other best practice for ACR in the cluster context.
```bash
kubectl check-acr foobar.azurecr.io
```

## Demo
[![asciicast](https://asciinema.org/a/373337.svg)](https://asciinema.org/a/373337)
