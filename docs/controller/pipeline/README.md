Pipeline Controller
=================

Pipeline controller is a controller for running scripts / playbooks on the server side.  It provides automation
capabilities to infrakit by making it possible to script playbook calls against a stream of targets as they become
available.

1. Install infrakit

```
$ curl -sSL https://docker.github.io/infrakit/install | sh
Building for mac darwin / amd64
Building infrakit GOOS=darwin GOARCH=amd64, version=f9310606.m, revision=f9310606c5e1ae6afa2d7d46ffbc110351da1e67, buildtags=builtin providers
```

2. Add this playbook
```
infrakit playbook add pipeline https://raw.githubusercontent.com/docker/infrakit/master/docs/controller/pipeline/playbook.yml
```

2. Verify playbook added
```
infrakit use pipeline
```

3. Start up server
```
infrakit use pipeline start --accept-defaults
```

4. View objects in the pipeline controller.  In another window

```
watch -d infrakit local mystack/pipeline describe -o
```

5. In another shell, commit the config yml:
```
infrakit use pipeline test1.yml
```
