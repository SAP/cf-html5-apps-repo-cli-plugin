package clients

import (
	"cf-html5-apps-repo-cli-plugin/log"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

var IsInsecure = false
var CustomCAPath = ""

func SetInsecure(isInsecure bool) {
	if isInsecure {
		log.Tracef("WARNING: SSL validation is disabled. To enable it login again using 'cf login', without '--skip-ssl-validation' flag\n")
	}
	IsInsecure = isInsecure
}

func SetCustomCAPath(customCAPath string) {
	if customCAPath != "" {
		log.Tracef("Custom CAs from %q are temporary added to system certificates pool. To use default system certificate pool, unset 'SSL_CERT_FILE' and 'SSL_CERT_DIR' environment variables before running CF CLI commands\n", customCAPath)
	}
	CustomCAPath = customCAPath
}

func GetDefaultClient() (client *http.Client, err error) {
	return GetClient(IsInsecure, CustomCAPath)
}

func GetClientWithCertificates(certificates []tls.Certificate) (client *http.Client, err error) {
	client, err = GetDefaultClient()
	if err == nil {
		(client.Transport.(*http.Transport)).TLSClientConfig.Certificates = append((client.Transport.(*http.Transport)).TLSClientConfig.Certificates, certificates...)
	}
	return
}

func GetClient(trustInsecure bool, customCAPath string) (client *http.Client, err error) {
	// No custom CA needed
	if customCAPath == "" {
		config := &tls.Config{InsecureSkipVerify: trustInsecure}
		tr := &http.Transport{TLSClientConfig: config}
		client = &http.Client{Transport: tr}
		return
	}

	// Get system certificates pool
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		return client, fmt.Errorf("reading system certificates failed: %s\n", err.Error())
	}

	// Initialize root CAs pool as empty, if no system CAs available
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	// Read local cert file
	certs, err := os.ReadFile(customCAPath)
	if err != nil {
		return client, fmt.Errorf("failed to append %q to RootCAs: %s\n", customCAPath, err.Error())
	}

	// Append custom CA to pool
	if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
		return client, fmt.Errorf("no certs appended, using system certs only\n")
	}

	// Trust additional certificate
	config := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: config}
	client = &http.Client{Transport: tr}

	return
}
