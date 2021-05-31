###############################################################################
# Spec file for fcfsfused-proxy
################################################################################
# Configured to be built by  non-root user
################################################################################
#
Summary: Utility scripts for creating RPM package for fcfsfused-proxy
Name: fcfsfused-proxy
Version: v0.1.0
Release: 1
License: Apache
Group: System
Packager: David Both
Requires: bash
BuildRoot: ~/rpmbuild/

%description
Utility scripts for creating RPM package for fcfsfused-proxy

%install
mkdir -p %{buildroot}/usr/bin/
cp fcfsfused-proxy %{buildroot}/usr/bin/fcfsfused-proxy

%files
/usr/bin/fcfsfused-proxy
