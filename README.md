# spiffe-workload-identity-demo

This demo has been created as part of this
[blog post](http://jetstack.io/blog/workload-identity-with-spiffe-trust-domains/),
please see that post for a detailed explanation of what this is all about.

This repo is a demo consisting of the following:

* set up for cert-manager, SPIFFE CSI driver, cert-manager trust on a kind cluster
* sample applications to communicate using SPIFFE flavoured mTLS

It should be possible to bring up the demo with `make demo`. You must have
these tools installed:

* kind
* kubectl
* cmctl
* goreleaser
* docker or docker compatible
