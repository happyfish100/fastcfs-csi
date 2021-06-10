/*
Copyright 2018 The Ceph-CSI Authors.

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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	adminName            = "adminName"
	adminSecretKey       = "adminSecretKey"
	userName             = "userName"
	userSecretKey        = "userSecretKey"
	tmpKeyFileLocation   = "/tmp/csi/keys"
	tmpKeyFileNamePrefix = "keyfile-"
)

// Credentials struct represents credentials to access the FastCFS cluster.
type Credentials struct {
	UserName string
	KeyFile  string
}

func storeKey(key string) (string, error) {
	tmpfile, err := ioutil.TempFile(tmpKeyFileLocation, tmpKeyFileNamePrefix)
	if err != nil {
		return "", fmt.Errorf("error creating a temporary keyfile: %w", err)
	}
	defer func() {
		if err != nil {
			// don't complain about unhandled error
			_ = os.Remove(tmpfile.Name())
		}
	}()

	if _, err = tmpfile.Write([]byte(key)); err != nil {
		return "", fmt.Errorf("error writing key to temporary keyfile: %w", err)
	}

	keyFile := tmpfile.Name()
	if keyFile == "" {
		err = fmt.Errorf("error reading temporary filename for key: %w", err)
		return "", err
	}

	if err = tmpfile.Close(); err != nil {
		return "", fmt.Errorf("error closing temporary filename: %w", err)
	}

	return keyFile, nil
}

func newCredentialsFromSecret(idField, keyField string, secrets map[string]string) (*Credentials, error) {
	var (
		c  = &Credentials{}
		ok bool
	)

	if len(secrets) == 0 {
		return nil, errors.New("provided secret is empty")
	}
	if c.UserName, ok = secrets[idField]; !ok {
		return nil, fmt.Errorf("missing ID field '%s' in secrets", idField)
	}

	key := secrets[keyField]
	if key == "" {
		return nil, fmt.Errorf("missing key field '%s' in secrets", keyField)
	}

	keyFile, err := storeKey(key)
	if err == nil {
		c.KeyFile = keyFile
	}

	return c, err
}

// DeleteCredentials removes the KeyFile.
func (cr *Credentials) DeleteCredentials() {
	// don't complain about unhandled error
	_ = os.Remove(cr.KeyFile)
}

// NewAdminCredentials creates new admin credentials from secret.
func NewAdminCredentials(secrets map[string]string) (*Credentials, error) {
	return newCredentialsFromSecret(adminName, adminSecretKey, secrets)
}

func NewUserCredentials(secrets map[string]string) (*Credentials, error) {
	return newCredentialsFromSecret(userName, userSecretKey, secrets)
}

func GetCredentialsForVolume(pre bool, secrets map[string]string) (*Credentials, error) {
	var (
		err error
		cr  *Credentials
	)

	if pre {
		// The volume is pre-made, credentials are in node stage secrets

		cr, err = NewUserCredentials(secrets)
		if err != nil {
			return nil, fmt.Errorf("failed to get user credentials from secrets: %w", err)
		}
	} else {
		// The volume is provisioned dynamically, use passed in admin credentials

		cr, err = NewAdminCredentials(secrets)
		if err != nil {
			return nil, fmt.Errorf("failed to get admin credentials from secrets: %w", err)
		}
	}

	return cr, nil
}
