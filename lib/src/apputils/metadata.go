// Copyright 2016 The Web BSD Hunt Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
////////////////////////////////////////////////////////////////////////////////
//
// TODO: High-level file comment.
package apputils

import(
	"encoding/json"
	"strings"

	"cloud.google.com/go/compute/metadata"
)

type AuthToken struct {
	Type	string	`json:"token_type"`
	Token	string	`json:"access_token"`
	Expiry	uint	`json:"expires_in"`
}

type ServiceAccount struct {
	Aliases	[]string	`json:"aliases"`
	Email	string		`json:"email"`
	Scopes	[]string	`json:"scopes"`
}

func GAEBackendInstance() (string, error) {
	return metadata.InstanceAttributeValue("gae_backend_instance")
}

func ServiceAccounts() ([]string, error) {
	return list("instance/service-accounts/", true)
}

func GetServiceAccount(name string) (*ServiceAccount, error) {
	s, err := metadata.Get("instance/service-accounts/" + name + "/")
	if err != nil {
		return nil, err
	}

	account := &ServiceAccount{}
	err = json.Unmarshal([]byte(s), account)
	if err != nil {
		return nil, err
	}

	return account, nil
}

func GetAuthToken(account string) (*AuthToken, error) {
	s, err := metadata.Get("instance/service-accounts/" + account + "/token")
	if err != nil {
		return nil, err
	}

	token := &AuthToken{}
	err = json.Unmarshal([]byte(s), token)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func list(suffix string, stripdir bool) ([]string, error) {
	s, err := metadata.Get(suffix)
	if err != nil {
		return nil, err
	}

	var list []string
	for _, l := range strings.Split(s, "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		if stripdir && l[len(l)-1] == '/' {
			l = l[:len(l)-1]
		}

		list = append(list, strings.TrimSpace(l))
	}

	return list, nil
}

func ProjectID() (string, error) {
	return metadata.ProjectID()
}
