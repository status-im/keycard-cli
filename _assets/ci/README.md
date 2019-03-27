# Description

These Jenkinsfiles are using for jobs under [`status-keycard`](https://ci.status.im/job/status-keycard/) folder.

# Builds

Available are:

* [PR Builds](https://ci.status.im/job/status-keycard/job/prs/job/keycard-cli/)
  - Use separate [`Jenkinsfile.pr`](./Jenkinsfile.pr)
  - Run `make test` and `go build`
  - Create only one `keycard` binary artifact
* [Manual Builds](https://ci.status.im/job/status-keycard/job/keycard-cli/)
  - Use separate [`Jenkinsfile`](./Jenkinsfile)
  - Build for 3 platforms using `xgo`
  - Create and replace GitHub release draft

# Known Issues

The manual builds remove the existing GitHub release and replease it with a new one. To avoid this update the [`VERSION`](/VERSION) file once you've created a reales you don't want removed.
