/*
Copyright 2020 The Ceph-CSI Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

/*
Modifications Copyright 2021 vazmin.
Licensed under the Apache License, Version 2.0.
*/

package common

import (
	"fmt"
	"github.com/container-storage-interface/spec/lib/go/csi"
	"k8s.io/klog/v2"
	"strings"
)

const (
	keySeparator   rune   = '/'
	labelSeparator string = ","
)


func GetTopologyFromParams(params map[string]string, requirement *csi.TopologyRequirement) map[string]string {
	domainLabels, exists := params["domainLabels"]
	if !exists {
		return nil
	}
	labels := parseDomainLabels(domainLabels)
	if requirement == nil {
		return nil
	}
	if topology, done := getTopology(labels, requirement.GetPreferred()); done {
		return topology
	}
	if topology, done := getTopology(labels, requirement.GetRequisite()); done {
		return topology
	}
	return nil
}

func getTopology(domainLabel []string, csiTopology []*csi.Topology) (map[string]string, bool) {
	topologyMap := make(map[string]string)
	for _, topology := range csiTopology {
		for _, label := range domainLabel {
			domain, exists := topology.GetSegments()[label]
			if exists {
				topologyMap[label] = domain
			}
		}
	}
	if len(topologyMap) > 0 {
		return topologyMap, true
	}
	return nil, false
}

func parseDomainLabels(domainLabels string) []string {
	return strings.SplitN(domainLabels, labelSeparator, -1)
}

// GetTopologyFromDomainLabels returns the CSI topology map, determined from
// the domain labels and their values from the CO system
// Expects domainLabels in arg to be in the format "[prefix/]<name>,[prefix/]<name>,...",.
func GetTopologyFromDomainLabels(nodeLabels map[string]string, domainLabels, driverName string) (map[string]string, error) {
	if domainLabels == "" {
		return nil, nil
	}

	// size checks on domain label prefix
	topologyPrefix := strings.ToLower("topology." + driverName)
	const lenLimit = 63
	if len(topologyPrefix) > lenLimit {
		return nil, fmt.Errorf("computed topology label prefix %q for node exceeds length limits", topologyPrefix)
	}
	// driverName is validated, and we are adding a lowercase "topology." to it, so no validation for conformance

	// Convert passed in labels to a map, and check for uniqueness
	labelsToRead := parseDomainLabels(domainLabels)
	klog.V(4).Infof("passed in node labels for processing: %+v", labelsToRead)

	labelsIn := make(map[string]bool)
	labelCount := 0
	for _, label := range labelsToRead {
		// as we read the labels from k8s, and check for missing labels,
		// no label conformance checks here
		if _, ok := labelsIn[label]; ok {
			return nil, fmt.Errorf("duplicate label %q found in domain labels", label)
		}

		labelsIn[label] = true
		labelCount++
	}

	// Determine values for requested labels from node labels
	domainMap := make(map[string]string)
	found := 0
	for key, value := range nodeLabels {
		if _, ok := labelsIn[key]; !ok {
			continue
		}
		// label found split name component and store value
		nameIdx := strings.IndexRune(key, keySeparator)
		domain := key[nameIdx+1:]
		domainMap[domain] = value
		labelsIn[key] = false
		found++
	}

	// Ensure all labels are found
	if found != labelCount {
		var missingLabels []string
		for key, missing := range labelsIn {
			if missing {
				missingLabels = append(missingLabels, key)
			}
		}
		return nil, fmt.Errorf("missing domain labels %v", missingLabels)
	}

	klog.V(4).Infof("list of domains processed: %+v", domainMap)

	topology := make(map[string]string)
	for domain, value := range domainMap {
		topology[topologyPrefix+"/"+domain] = value
		// TODO: when implementing domain takeover/giveback, enable a domain value that can remain pinned to the node
		// topology["topology."+driverName+"/"+domain+"-pinned"] = value
	}

	return topology, nil
}



