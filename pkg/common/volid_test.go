/*
Copyright 2021 vazmin.

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

package common

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestComposeCSIID(t *testing.T) {
	cid := &CSIIdentifier{
		ClusterID: "1",
		userName:  "admin",
		volName:   csiVolNamingPrefix + "pvc-4cb82c80-c1e9-4491-8625-e24b54dabb49",
	}

	csiid, err := cid.composeCSIID()
	if err != nil {
		t.Errorf(err.Error())

	}
	fmt.Println(csiid)

	decid := &CSIIdentifier{}
	decid.decomposeCSIID(csiid)
	assert.Equal(t, cid.ClusterID, decid.ClusterID)
	assert.Equal(t, cid.userName, decid.userName)
	assert.Equal(t, cid.volName, decid.volName)
}
