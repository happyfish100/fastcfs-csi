package driver

import (
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog/v2"
	"os"
)

//secretParamsMap provides a mapping of current as well as deprecated secret keys
type secretParamsMap struct {
	name                         string
	deprecatedSecretNameKey      string
	deprecatedSecretNamespaceKey string
	secretNameKey                string
	secretNamespaceKey           string
}

const (
	// CSI Parameters prefixed with csiParameterPrefix are not passed through
	// to the driver on CreateVolumeRequest calls. Instead they are intended
	// to used by the CSI external-provisioner and maybe used to populate
	// fields in subsequent CSI calls or Kubernetes API objects.
	csiParameterPrefix = "csi.storage.k8s.io/"

	prefixedFsTypeKey = csiParameterPrefix + "fstype"

	prefixedDefaultSecretNameKey      = csiParameterPrefix + "secret-name"
	prefixedDefaultSecretNamespaceKey = csiParameterPrefix + "secret-namespace"

	prefixedProvisionerSecretNameKey      = csiParameterPrefix + "provisioner-secret-name"
	prefixedProvisionerSecretNamespaceKey = csiParameterPrefix + "provisioner-secret-namespace"

	prefixedControllerPublishSecretNameKey      = csiParameterPrefix + "controller-publish-secret-name"
	prefixedControllerPublishSecretNamespaceKey = csiParameterPrefix + "controller-publish-secret-namespace"

	prefixedNodeStageSecretNameKey      = csiParameterPrefix + "node-stage-secret-name"
	prefixedNodeStageSecretNamespaceKey = csiParameterPrefix + "node-stage-secret-namespace"

	prefixedNodePublishSecretNameKey      = csiParameterPrefix + "node-publish-secret-name"
	prefixedNodePublishSecretNamespaceKey = csiParameterPrefix + "node-publish-secret-namespace"

	prefixedControllerExpandSecretNameKey      = csiParameterPrefix + "controller-expand-secret-name"
	prefixedControllerExpandSecretNamespaceKey = csiParameterPrefix + "controller-expand-secret-namespace"

	// [Deprecated] CSI Parameters that are put into fields but
	// NOT stripped from the parameters passed to CreateVolume
	provisionerSecretNameKey      = "csiProvisionerSecretName"
	provisionerSecretNamespaceKey = "csiProvisionerSecretNamespace"

	controllerPublishSecretNameKey      = "csiControllerPublishSecretName"
	controllerPublishSecretNamespaceKey = "csiControllerPublishSecretNamespace"

	nodeStageSecretNameKey      = "csiNodeStageSecretName"
	nodeStageSecretNamespaceKey = "csiNodeStageSecretNamespace"

	nodePublishSecretNameKey      = "csiNodePublishSecretName"
	nodePublishSecretNamespaceKey = "csiNodePublishSecretNamespace"

	// PV and PVC metadata, used for sending to drivers in the  create requests, added as parameters, optional.
	pvcNameKey      = "csi.storage.k8s.io/pvc/name"
	pvcNamespaceKey = "csi.storage.k8s.io/pvc/namespace"
	pvNameKey       = "csi.storage.k8s.io/pv/name"

	pvcKind = "PersistentVolumeClaim" // Native types don't require an API group

	tokenPVNameKey       = "pv.name"
	tokenPVCNameKey      = "pvc.name"
	tokenPVCNameSpaceKey = "pvc.namespace"

	pvcCloneFinalizer = "provisioner.storage.kubernetes.io/cloning-protection"
)

var (
	defaultSecretParams = secretParamsMap{
		name:               "Default",
		secretNameKey:      prefixedDefaultSecretNameKey,
		secretNamespaceKey: prefixedDefaultSecretNamespaceKey,
	}

	nodeStageSecretParams = secretParamsMap{
		name:                         "NodeStage",
		deprecatedSecretNameKey:      nodeStageSecretNameKey,
		deprecatedSecretNamespaceKey: nodeStageSecretNamespaceKey,
		secretNameKey:                prefixedNodeStageSecretNameKey,
		secretNamespaceKey:           prefixedNodeStageSecretNamespaceKey,
	}
)

func getNodeStageSecretReference(storageClassParams map[string]string, pvName string, pvc *corev1.PersistentVolumeClaim) (*corev1.SecretReference, error) {
	return getSecretReference(nodeStageSecretParams, storageClassParams, pvName, pvc)
}

// getSecretReference returns a reference to the secret specified in the given nameTemplate
//  and namespaceTemplate, or an error if the templates are not specified correctly.
// no lookup of the referenced secret is performed, and the secret may or may not exist.
//
// supported tokens for name resolution:
// - ${pv.name}
// - ${pvc.namespace}
// - ${pvc.name}
// - ${pvc.annotations['ANNOTATION_KEY']} (e.g. ${pvc.annotations['example.com/node-publish-secret-name']})
//
// supported tokens for namespace resolution:
// - ${pv.name}
// - ${pvc.namespace}
//
// an error is returned in the following situations:
// - the nameTemplate or namespaceTemplate contains a token that cannot be resolved
// - the resolved name is not a valid secret name
// - the resolved namespace is not a valid namespace name
func getSecretReference(secretParams secretParamsMap, storageClassParams map[string]string, pvName string, pvc *corev1.PersistentVolumeClaim) (*corev1.SecretReference, error) {
	nameTemplate, namespaceTemplate, err := verifyAndGetSecretNameAndNamespaceTemplate(secretParams, storageClassParams)
	if err != nil {
		return nil, fmt.Errorf("failed to get name and namespace template from params: %v", err)
	}

	// if didn't find secrets for specific call, try to check default values
	if nameTemplate == "" && namespaceTemplate == "" {
		nameTemplate, namespaceTemplate, err = verifyAndGetSecretNameAndNamespaceTemplate(defaultSecretParams, storageClassParams)
		if err != nil {
			return nil, fmt.Errorf("failed to get default name and namespace template from params: %v", err)
		}
	}

	if nameTemplate == "" && namespaceTemplate == "" {
		return nil, nil
	}

	ref := &corev1.SecretReference{}
	{
		// Secret namespace template can make use of the PV name or the PVC namespace.
		// Note that neither of those things are under the control of the PVC user.
		namespaceParams := map[string]string{tokenPVNameKey: pvName}
		if pvc != nil {
			namespaceParams[tokenPVCNameSpaceKey] = pvc.Namespace
		}

		resolvedNamespace, err := resolveTemplate(namespaceTemplate, namespaceParams)
		if err != nil {
			return nil, fmt.Errorf("error resolving value %q: %v", namespaceTemplate, err)
		}
		if len(validation.IsDNS1123Label(resolvedNamespace)) > 0 {
			if namespaceTemplate != resolvedNamespace {
				return nil, fmt.Errorf("%q resolved to %q which is not a valid namespace name", namespaceTemplate, resolvedNamespace)
			}
			return nil, fmt.Errorf("%q is not a valid namespace name", namespaceTemplate)
		}
		ref.Namespace = resolvedNamespace
	}

	{
		// Secret name template can make use of the PV name, PVC name or namespace, or a PVC annotation.
		// Note that PVC name and annotations are under the PVC user's control.
		nameParams := map[string]string{tokenPVNameKey: pvName}
		if pvc != nil {
			nameParams[tokenPVCNameKey] = pvc.Name
			nameParams[tokenPVCNameSpaceKey] = pvc.Namespace
			for k, v := range pvc.Annotations {
				nameParams["pvc.annotations['"+k+"']"] = v
			}
		}
		resolvedName, err := resolveTemplate(nameTemplate, nameParams)
		if err != nil {
			return nil, fmt.Errorf("error resolving value %q: %v", nameTemplate, err)
		}
		if len(validation.IsDNS1123Subdomain(resolvedName)) > 0 {
			if nameTemplate != resolvedName {
				return nil, fmt.Errorf("%q resolved to %q which is not a valid secret name", nameTemplate, resolvedName)
			}
			return nil, fmt.Errorf("%q is not a valid secret name", nameTemplate)
		}
		ref.Name = resolvedName
	}

	return ref, nil
}

func resolveTemplate(template string, params map[string]string) (string, error) {
	missingParams := sets.NewString()
	resolved := os.Expand(template, func(k string) string {
		v, ok := params[k]
		if !ok {
			missingParams.Insert(k)
		}
		return v
	})
	if missingParams.Len() > 0 {
		return "", fmt.Errorf("invalid tokens: %q", missingParams.List())
	}
	return resolved, nil
}

// verifyAndGetSecretNameAndNamespaceTemplate gets the values (templates) associated
// with the parameters specified in "secret" and verifies that they are specified correctly.
func verifyAndGetSecretNameAndNamespaceTemplate(secret secretParamsMap, storageClassParams map[string]string) (nameTemplate, namespaceTemplate string, err error) {
	numName := 0
	numNamespace := 0

	if t, ok := storageClassParams[secret.deprecatedSecretNameKey]; ok {
		nameTemplate = t
		numName++
		klog.Warning(deprecationWarning(secret.deprecatedSecretNameKey, secret.secretNameKey, ""))
	}
	if t, ok := storageClassParams[secret.deprecatedSecretNamespaceKey]; ok {
		namespaceTemplate = t
		numNamespace++
		klog.Warning(deprecationWarning(secret.deprecatedSecretNamespaceKey, secret.secretNamespaceKey, ""))
	}
	if t, ok := storageClassParams[secret.secretNameKey]; ok {
		nameTemplate = t
		numName++
	}
	if t, ok := storageClassParams[secret.secretNamespaceKey]; ok {
		namespaceTemplate = t
		numNamespace++
	}

	if numName > 1 || numNamespace > 1 {
		// Double specified error
		return "", "", fmt.Errorf("%s secrets specified in parameters with both \"csi\" and \"%s\" keys", secret.name, csiParameterPrefix)
	} else if numName != numNamespace {
		// Not both 0 or both 1
		return "", "", fmt.Errorf("either name and namespace for %s secrets specified, Both must be specified", secret.name)
	} else if numName == 1 {
		// Case where we've found a name and a namespace template
		if nameTemplate == "" || namespaceTemplate == "" {
			return "", "", fmt.Errorf("%s secrets specified in parameters but value of either namespace or name is empty", secret.name)
		}
		return nameTemplate, namespaceTemplate, nil
	} else if numName == 0 {
		// No secrets specified
		return "", "", nil
	} else {
		// THIS IS NOT A VALID CASE
		return "", "", fmt.Errorf("unknown error with getting secret name and namespace templates")
	}
}

func deprecationWarning(deprecatedParam, newParam, removalVersion string) string {
	if removalVersion == "" {
		removalVersion = "a future release"
	}
	newParamPhrase := ""
	if len(newParam) != 0 {
		newParamPhrase = fmt.Sprintf(", please use \"%s\" instead", newParam)
	}
	return fmt.Sprintf("\"%s\" is deprecated and will be removed in %s%s", deprecatedParam, removalVersion, newParamPhrase)
}
