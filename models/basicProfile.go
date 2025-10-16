/*
Copyright 2023 The Cloud-Barista Authors.

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
package models

import "mime/multipart"

type BaseProfile struct {
	ProfileName string `json:"profileName" form:"profileName"`
}

type PublicKeyResponse struct {
	PublicKeyTokenId string `json:"publicKeyTokenId"`
	PublicKey        string `json:"publicKey"`
}

type TumblebugCredentialCreate struct {
	CredentialHolder                 string            `json:"credentialHolder"`
	CredentialKeyValueList           map[string]string `json:"credentialKeyValueList"`
	EncryptedClientAesKeyByPublicKey string            `json:"encryptedClientAesKeyByPublicKey"`
	ProviderName                     string            `json:"providerName"`
	PublicKeyTokenId                 string            `json:"publicKeyTokenId"`
}

type ProfileCredentials struct {
	AWS AWSCredentials `json:"aws,omitempty"`
	NCP NCPCredentials `json:"ncp,omitempty"`
	GCP GCPCredentials `json:"gcp,omitempty"`
}

type AWSCredentials struct {
	AccessKey string `json:"accessKey" form:"accessKey"`
	SecretKey string `json:"secretKey" form:"secretKey"`
}

type NCPCredentials struct {
	AccessKey string `json:"accessKey" form:"accessKey"`
	SecretKey string `json:"secretKey" form:"secretKey"`
}

type GCPCredentials struct {
	Type                string `json:"type" form:"type"`
	ProjectID           string `json:"project_id" form:"project_id"`
	PrivateKeyID        string `json:"private_key_id" form:"private_key_id"`
	PrivateKey          string `json:"private_key" form:"private_key"`
	ClientEmail         string `json:"client_email" form:"client_email"`
	ClientID            string `json:"client_id" form:"client_id"`
	AuthURI             string `json:"auth_uri" form:"auth_uri"`
	TokenURI            string `json:"token_uri" form:"token_uri"`
	AuthProviderCertURL string `json:"auth_provider_x509_cert_url" form:"auth_provider_x509_cert_url"`
	ClientCertURL       string `json:"client_x509_cert_url" form:"client_x509_cert_url"`
	UniverseDomain      string `json:"universe_domain" form:"universe_domain"`
}

type GCPCredentalCreateParams struct {
	GCPCredentialJson string                `form:"gcpCredentialJson" json:"gcpCredentialJson"`
	GCPCredential     *multipart.FileHeader `form:"gcpCredential" json:"-" swaggerignore:"true"`
}
