/*
 * Copyright (C) 2025 Gloria Ciavarrini
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package korifi

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

// Load Kubernetes configuration from the kubeconfig file
func getKubeConfig() (*api.Config, error) {
	home := homedir.HomeDir()
	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %v", err)
	}
	return config, nil
}

// Extract client certificate and key for authentication
func getPEMCertificate(config *api.Config) (string, error) {
	var dataCert, keyCert []byte

	for username, authInfo := range config.AuthInfos {
		if username == "kind-korifi" {
			dataCert = authInfo.ClientCertificateData
			keyCert = authInfo.ClientKeyData
			break
		}
	}

	if len(dataCert) == 0 || len(keyCert) == 0 {
		return "", fmt.Errorf("could not find certificate data for kind-korifi")
	}

	return base64.StdEncoding.EncodeToString(append(dataCert, keyCert...)), nil
}

// Custom RoundTripper for injecting Authorization headers
type authHeaderRoundTripper struct {
	certPEM string
	base    http.RoundTripper
}

func (t *authHeaderRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())
	reqClone.Header.Set("Authorization", "ClientCert "+t.certPEM)
	return t.base.RoundTrip(reqClone)
}

// Create an HTTP client configured to authenticate with Korifi
func GetKorifiHttpClient() (*http.Client, error) {
	kubeConfig, err := getKubeConfig()
	if err != nil {
		return nil, err
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Use with caution in production environments
		},
	}
	certPEM, err := getPEMCertificate(kubeConfig)
	if err != nil {
		return nil, err
	}

	roundTripper := &authHeaderRoundTripper{
		certPEM: certPEM,
		base:    transport,
	}

	return &http.Client{
		Transport: roundTripper,
	}, nil
}

// Define InfoV3Response struct based on API response fields.
type InfoV3Response struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Fetch information from the Korifi `/v3/info` endpoint
func GetInfo(httpClient *http.Client) (*InfoV3Response, error) {
	resp, err := httpClient.Get("https://localhost/v3/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	var info InfoV3Response

	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling info: %w", err)
	}

	return &info, nil
}
