# bookstore-sample-controller

A Controller written in kubernetes sample-controller style which watches a custom resource named Bookstore.
A resource create a deployment and a NodePort service. The container image is a simple
[bookstore api server](https://github.com/Shaad7/bookstore-api-server). The NodePort service listen to
request on host port 30000.

## How to Use

Clone repo and move to directory
```shell
git clone https://github.com/Shaad7/bookstore-sample-controller
cd bookstore-sample-controller
```

Create a cluster using  [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
```shell
kind create cluster --config=clusterconfig.yaml 
```

Create bookstores.gopher.com Custom Resource Definition (CRD)
```shell
kubectl apply -f artifacts/gopher.com_bookstore-status-subresource.yaml
```

Build and Run the controller
```shell
 go build -o bin/bookstore-controller .
 ./bin/bookstore-controller -kubeconfig=$HOME/.kube/config 
```

Create Custom Resource
```shell
kubectl apply -f artifacts/sample-bookstore.yaml 
```

Now the bookstore api server listens and serves to host port 30000.
To get list of the books
```shell
curl http://localhost:30000
```

All the API calls from [this list](https://github.com/Shaad7/bookstore-api-server#api-calls)
can be done on `http://localhost:30000` instead of `http://localhost:8081`
