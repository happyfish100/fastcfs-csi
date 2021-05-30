# fastcfs-csi

A Container Storage Interface (CSI) Driver for FastCFS.
The CSI plugin allows you to use FastCFS with your preferred Container Orchestrator.
It allows dynamically provisioning FastCFS volumes and attaching them to workloads.

## Project status

Status: **Alpha**

## How to start

see [deploy](./docs/deploy.md)

## Support

The driver is currently developed with csi spec v1.4.0, and tested on kubernetes v1.20+.

### CSI spec and Kubernetes version compatibility

Please refer to the [matrix](https://kubernetes-csi.github.io/docs/#kubernetes-releases)
in the Kubernetes documentation.
