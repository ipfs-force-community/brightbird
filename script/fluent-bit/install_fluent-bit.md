
```sh
 helm install -f ./values.yaml fluent-bit fluent/fluent-bit
```

```forward
export POD_NAME=$(kubectl get pods --platform default -l "app.kubernetes.io/name=fluent-bit,app.kubernetes.io/instance=fluent-bit" -o jsonpath="{.items[0].metadata.name}")
kubectl --namespace platform port-forward $POD_NAME 2020:2020
curl http://127.0.0.1:2020
```