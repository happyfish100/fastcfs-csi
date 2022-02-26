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
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	maxVolIDLen = 128
)

type CSIIdentifier struct {
	ClusterID string
	UserName  string
	VolName   string
}

type CSIIdentifierDecompose struct {
	composedCSIID string
	cursor        uint16
}

func (cid *CSIIdentifier) BasePath() string {
	return cid.ClusterID
}

func (cid *CSIIdentifier) Len() int {
	return 5*3 + len(cid.ClusterID) + len(cid.UserName) + len(cid.VolName)
}

// ComposeCSIID
//      [length of ClusterID=1:4byte] + [-:1byte]
//   	[ClusterID] + [-:1byte]
//  	[length of userName=1:4byte] + [-:1byte]
//   	[userName]
//  	[length of volName=1:4byte] + [-:1byte]
//   	[volName]
func (cid *CSIIdentifier) ComposeCSIID() (string, error) {
	buf16 := make([]byte, 2)

	if (cid.Len()) > maxVolIDLen {
		return "", fmt.Errorf("CSI ID encoding length overflow")
	}

	binary.BigEndian.PutUint16(buf16, uint16(len(cid.ClusterID)))
	clusterIDLength := hex.EncodeToString(buf16)

	binary.BigEndian.PutUint16(buf16, uint16(len(cid.UserName)))
	userNameLength := hex.EncodeToString(buf16)

	binary.BigEndian.PutUint16(buf16, uint16(len(cid.VolName)))
	volNameLength := hex.EncodeToString(buf16)

	return strings.Join([]string{clusterIDLength, cid.ClusterID, userNameLength, cid.UserName, volNameLength, cid.VolName}, "-"), nil
}

func (cid *CSIIdentifier) DecomposeCSIID(composedCSIID string) (err error) {

	cidd := &CSIIdentifierDecompose{
		composedCSIID: composedCSIID,
		cursor:        uint16(0),
	}

	cid.ClusterID, err = cidd.next()
	if err != nil {
		return err
	}

	cid.UserName, err = cidd.next()
	if err != nil {
		return err
	}

	cid.VolName, err = cidd.next()

	return err
}

func (cidd *CSIIdentifierDecompose) next() (string, error) {
	buf16, err := hex.DecodeString(cidd.composedCSIID[cidd.cursor : cidd.cursor+4])
	if err != nil {
		return "", err
	}
	clusterIDLength := binary.BigEndian.Uint16(buf16)
	cidd.cursor += 5
	end := cidd.cursor + clusterIDLength
	s := cidd.composedCSIID[cidd.cursor:end]
	cidd.cursor = end + 1
	return s, nil
}
