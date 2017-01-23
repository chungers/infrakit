<!--[metadata]>
+++
title = "Open Source"
description = "Docker for GCP Open Source"
keywords = ["iaas, gcp, oss"]
[menu.main]
identifier="faq-oss"
parent = "docs-gcp-faq"
name = "Open Source"
weight="150"
+++
<![end-metadata]-->

# Open Source

Docker for AWS and Azure Editions are built using open source software.

Docker for AWS and Azure Editions distribute some components that are licensed under the GNU General Public License. You can download the source for these components [here](https://download.docker.com/opensource/License.tar.gz).

The sources for qemu-img can be obtained [here](http://wiki.qemu-project.org/download/qemu-2.4.1.tar.bz2). The sources for the gettext and glib libraries that qemu-img requires were obtained from [Homebrew](https://brew.sh/) and may be retrieved using `brew install --build-from-source gettext glib`.