# $${\color{white}Config \space Keys \space Operator}$$

> [!CAUTION]
> Project is still in active development

Config Keys Operator can be used to monitor the keys in a configmap. A CRD of kind `EnvKeyMonitor` and API group `config.core.nvsh-ram.io` can be used to achieve this functionality.

Secrets are most commonly used for sensitive date in a cluster as best practice. This is because configmaps are stored on disk where as secrets are stored in memory. To enforce this, this operator can be used. 

## Table of Contents

- [How to Use](#how-to-use)
- [Functionality](#functionality)
    [Notes](#notes)
- [EnvKeyMonitor](#envkeymonitor)
    [Manifest Definition](#manifest-definition)
- [Limitations](#limitations)
- [Status](#status)

## How to Use

Following steps can be used:

- Create an `EnvKeyMonitor` object in the desired namespace (see [envkeymonitor](#envkeymonitor) section)


## Functionality

- `EnvKeyMonitor` has a list of keys under `.spec.keys` that are forbidden in configmaps

- Any configmap created that has a key under `.spec.data{}` listed in an `EnvKeyMonitor` in the same namespace will not be created
    - the only way to create the configmap is by removing the forbidden keys

### Notes

- When creating a new `EnvKeyMonitor`, duplicate keys are automatically removed, a key is considered a duplicate if:
    - the key appears multiple times in the same `EnvKeyMonitor` multiple times
    - the key is already being monitored by another `EnvKeyMonitor` object in the same namespace

## EnvKeyMonitor

```yml
apiVersion: config.core.nvsh-ram.io/v1
kind: EnvKeyMonitor
metadata:
  name: <name>
  namespace: <namespace>
spec:
  keys:
    - API_KEY
  policy: PERMISSIVE
```

### Manifest definition

`.spec`
| Key  | Type  | Note  |
|:---:|:---:|:---:|
| keys  | `[]string`  | list of strings to monitor, case sensitive, min=1 max=25  |
| policy  | `PERMISSIVE` or `STRICT`  | *  |

\*  functionality not yet implemented

## Limitations

- Environmental variables mounted directly into pods, deployments, statefulsets etc. are not monitored

## Status