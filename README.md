# artillery-operator

## Sample LoadTest CR

```yaml
apiVersion: loadtest.artillery.io/v1alpha1
kind: LoadTest
metadata:
  name: loadtest-sample
  namespace: default
spec:
  # Add fields here
  count: 10
  environment: stage
  testScript:
    config:
      configMap: load-test-config
    external:
      payload:
        configMaps:
          - csv-payload-1
          - csv-payload-2
          - csv-payload-3
      processor:
        main:
          configMap: my-functions-js
        related:
          configMaps:
            - package-json
            - helper-js
```

More setup info coming soon.
