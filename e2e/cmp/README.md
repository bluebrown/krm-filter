# ArgoCD CMP

## Authentication

### Github Token

Put your github email and token in e2e/cmp/deploy/etc/git-creds.env

```bash
username=user.com
password=ghp_foobarbaz
```

### Azure Vault Token

The token will be automatically fetched, but not renewed. If its expired, delete
e2e/cmp/deploy/components/aad-token/etc/aad-token.env to to get a fresh one.

## Deploy

Once you have taken care of the credentials, you can run a local kind cluster to test test CMP.

```bash
# create local cluster
make kind

# deploy argocd with cmp
make deploy

# port forward in the background
nohup kubectl port-forward svc/argocd-server -n argocd 8080:443 &

# get the admin password to login on http://localhost:8080
kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d; echo

# deploy test apps, testing the cmp
kubectl apply -f e2e/cmp/testdata/kustomize/
kubectl apply -f e2e/cmp/testdata/kpt/
```
