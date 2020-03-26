# Prometheus Alert Overrider

Tool for overriding default prometheus alerts.
This tool is designed to work as *ansible module*.

# Usage

The application loads all of the valid files from specified directory and applies overrides if they are defined.

```

# Override all rules beginning with K8S
groups:
- name: k8s-workload-container-overrides
  rules:
  - alert: DisableKubeDev
    override: ["K8S.*"] # List of rules to be overriden, accepts regexp
    # If set to false, only default rules are changes and no new rules are created
    enabled: false
    # This expresion will be negated and inserted into all rules matching the values in ovveride. If enabled is set to true, use this field to define query for new rule. Only filters will be negated and inserted into default rules
    expr: '{kubernetes_cluster="kube-prod"}'
```

Only alert rules are altered, recording rules are left intact and are just passed to output without change.


Running the program
```
./prometheus_merget <path_to_rules>
```

## Build
```
CGO_ENABLED=0 GOOS=linux go build -o prometheus_alert_overrider main.go
```