# Turtles integration suite example

This repository contains an example of how to verify the Rancher Turtles integration of CAPI providers. At the moment we only require one test that uses a GitOps flow.
Running this suite for a CAPI provider will:
- Create a managment cluster in the desired environment.
- Install Rancher and Turtles with all prerequisites.
- Install Gitea.
- Run the suite that will create a git repo, apply cluster template using Fleet and verify the cluster is created and successfully imported in Rancher.

## Running the suite

### Prerequisites

Please follow the instructions in the [Rancher Turtles repository](https://github.com/rancher/turtles/tree/main/test/e2e#e2e-tests) to set up the environment.

### Running the suite

To run the suite, execute the following command:

```bash
make test
```
