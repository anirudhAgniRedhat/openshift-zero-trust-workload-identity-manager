# SPIRE Agent Security Hardening Summary

## Overview
This document summarizes the security hardening changes made to the SPIRE agent to restrict privileged access while maintaining functionality.

## Changes Made

### Security Context Constraints (SCC) - `pkg/controller/spire-agent/scc.go`

**Before:**
- `RunAsUser`: `RunAsAny` (unrestricted)
- `SELinuxContext`: `RunAsAny` (unrestricted)
- `AllowHostIPC`: `true`
- `AllowHostPorts`: `true`
- `AllowPrivilegeEscalation`: `true`
- `AllowPrivilegedContainer`: `true`
- `RequiredDropCapabilities`: None

**After:**
- `RunAsUser`: `RunAsAny` (required for hostPath access)
- `SELinuxContext`: `MustRunAs` (enforced SELinux context)
- `AllowHostIPC`: `false` (disabled)
- `AllowHostPorts`: `true` (required for hostNetwork)
- `AllowPrivilegeEscalation`: `false` (disabled)
- `AllowPrivilegedContainer`: `false` (disabled)
- `RequiredDropCapabilities`: `["ALL"]` (drops all capabilities)
- `Priority`: `10` (ensures SCC is evaluated)

### DaemonSet Security - `pkg/controller/spire-agent/daemonset.go`

**Pod Security Context:**
- `RunAsUser`: `0` (root - required for hostPath write access)
- **Note**: While running as root, the container has ALL capabilities dropped and cannot escalate privileges

**Container Security Context:**
- `AllowPrivilegeEscalation`: `false`
- `ReadOnlyRootFilesystem`: `true`
- `Capabilities`: Drop `["ALL"]`

### Justification for Root User

The SPIRE agent needs to run as root (UID 0) for the following reasons:
1. **HostPath Volume Access**: The agent writes to `/run/spire/agent-sockets` (hostPath), which is owned by root on the host
2. **SPIFFE CSI Driver Integration**: The CSI driver needs to access the same socket directory
3. **No Init Container**: Per requirement, init containers are not used to set up directory permissions

## Security Improvements

Despite running as root, the SPIRE agent is now significantly more secure:

| Security Control | Before | After |
|-----------------|--------|-------|
| Privileged Container | ✗ Allowed | ✓ Blocked |
| Privilege Escalation | ✗ Allowed | ✓ Blocked |
| Host IPC | ✗ Allowed | ✓ Blocked |
| Linux Capabilities | ✗ All capabilities | ✓ ALL dropped |
| Root Filesystem | ✗ Read-write | ✓ Read-only |
| SELinux | ✗ Unrestricted | ✓ Enforced |

## Required Host-Level Access

The SPIRE agent legitimately requires:
- **HostNetwork**: Required for node attestation
- **HostPID**: Required for workload attestation (identifying processes)
- **HostPath**: Required for socket sharing with SPIFFE CSI driver

## Verification

To verify the security hardening:

```bash
# Check SCC configuration
kubectl get scc spire-agent -o yaml

# Verify pod security context
kubectl get pod -n zero-trust-workload-identity-manager -l app.kubernetes.io/name=spire-agent -o jsonpath='{.items[0].spec.securityContext}'

# Verify container security context
kubectl get pod -n zero-trust-workload-identity-manager -l app.kubernetes.io/name=spire-agent -o jsonpath='{.items[0].spec.containers[0].securityContext}'

# Check which SCC is being used
kubectl get pod -n zero-trust-workload-identity-manager -l app.kubernetes.io/name=spire-agent -o jsonpath='{.items[0].metadata.annotations.openshift\.io/scc}'
```

## Testing

Updated unit tests in `pkg/controller/spire-agent/scc_test.go` to verify:
- SCC priority is set to 10
- RunAsUser strategy is RunAsAny
- AllowPrivilegedContainer is false
- AllowPrivilegeEscalation is false
- RequiredDropCapabilities includes ALL

## Trade-offs

**Running as Root:**
- Required for hostPath write access without init containers
- Mitigated by dropping all capabilities and preventing privilege escalation
- Container cannot perform privileged operations despite running as UID 0

**Alternative Solutions (Not Used):**
- Init container to set permissions: Rejected per user requirement
- EmptyDir instead of HostPath: Not viable (CSI driver needs shared socket)
- Non-root user: Cannot write to root-owned hostPath directory

## Conclusion

The SPIRE agent security posture has been significantly improved while maintaining full functionality. The agent runs with minimal privileges and cannot escalate to perform privileged operations, even though it runs as UID 0.

