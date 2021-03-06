// Copyright 2018 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package topdown

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/topdown/builtins"
	"github.com/open-policy-agent/opa/util"
)

func builtinCryptoX509ParseCertificates(a ast.Value) (ast.Value, error) {

	str, err := builtinBase64Decode(a)
	if err != nil {
		return nil, err
	}

	certs, err := x509.ParseCertificates([]byte(str.(ast.String)))
	if err != nil {
		return nil, err
	}

	bs, err := json.Marshal(certs)
	if err != nil {
		return nil, err
	}

	var x interface{}

	if err := util.UnmarshalJSON(bs, &x); err != nil {
		return nil, err
	}

	return ast.InterfaceToValue(x)
}

func hashHelper(a ast.Value, h func(ast.String) string) (ast.Value, error) {
	s, err := builtins.StringOperand(a, 1)
	if err != nil {
		return nil, err
	}
	return ast.String(h(s)), nil
}

func builtinCryptoMd5(a ast.Value) (ast.Value, error) {
	return hashHelper(a, func(s ast.String) string { return fmt.Sprintf("%x", md5.Sum([]byte(s))) })
}

func builtinCryptoSha1(a ast.Value) (ast.Value, error) {
	return hashHelper(a, func(s ast.String) string { return fmt.Sprintf("%x", sha1.Sum([]byte(s))) })
}

func builtinCryptoSha256(a ast.Value) (ast.Value, error) {
	return hashHelper(a, func(s ast.String) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) })
}

func builtinCryptoX509ParseCertificateRequest(a ast.Value) (ast.Value, error) {
	str, err := builtinBase64Decode(a)
	if err != nil {
		return nil, err
	}

	p, _ := pem.Decode([]byte(str.(ast.String)))
	if p == nil || p.Type != "CERTIFICATE REQUEST" {
		return nil, fmt.Errorf("invalid PEM-encoded certificate signing request")
	}

	csr, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return nil, err
	}

	bs, err := json.Marshal(csr)
	if err != nil {
		return nil, err
	}

	var x interface{}
	if err := util.UnmarshalJSON(bs, &x); err != nil {
		return nil, err
	}
	return ast.InterfaceToValue(x)
}

func init() {
	RegisterFunctionalBuiltin1(ast.CryptoX509ParseCertificates.Name, builtinCryptoX509ParseCertificates)
	RegisterFunctionalBuiltin1(ast.CryptoMd5.Name, builtinCryptoMd5)
	RegisterFunctionalBuiltin1(ast.CryptoSha1.Name, builtinCryptoSha1)
	RegisterFunctionalBuiltin1(ast.CryptoSha256.Name, builtinCryptoSha256)
	RegisterFunctionalBuiltin1(ast.CryptoX509ParseCertificateRequest.Name, builtinCryptoX509ParseCertificateRequest)
}

// addCACertsFromFile adds CA certificates from filePath into the given pool.
// If pool is nil, it creates a new x509.CertPool. pool is returned.
func addCACertsFromFile(pool *x509.CertPool, filePath string) (*x509.CertPool, error) {
	if pool == nil {
		pool = x509.NewCertPool()
	}

	caCert, err := readCertFromFile(filePath)
	if err != nil {
		return nil, err
	}

	if ok := pool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("could not append CA certificates from %q", filePath)
	}

	return pool, nil
}

// addCACertsFromBytes adds CA certificates from pemBytes into the given pool.
// If pool is nil, it creates a new x509.CertPool. pool is returned.
func addCACertsFromBytes(pool *x509.CertPool, pemBytes []byte) (*x509.CertPool, error) {
	if pool == nil {
		pool = x509.NewCertPool()
	}

	if ok := pool.AppendCertsFromPEM(pemBytes); !ok {
		return nil, fmt.Errorf("could not append certificates")
	}

	return pool, nil
}

// addCACertsFromBytes adds CA certificates from the environment variable named
// by envName into the given pool. If pool is nil, it creates a new x509.CertPool.
// pool is returned.
func addCACertsFromEnv(pool *x509.CertPool, envName string) (*x509.CertPool, error) {
	pool, err := addCACertsFromBytes(pool, []byte(os.Getenv(envName)))
	if err != nil {
		return nil, fmt.Errorf("could not add CA certificates from envvar %q: %w", envName, err)
	}

	return pool, err
}

// ReadCertFromFile reads a cert from file
func readCertFromFile(localCertFile string) ([]byte, error) {
	// Read in the cert file
	certPEM, err := ioutil.ReadFile(localCertFile)
	if err != nil {
		return nil, err
	}
	return certPEM, nil
}

// ReadKeyFromFile reads a key from file
func readKeyFromFile(localKeyFile string) ([]byte, error) {
	// Read in the cert file
	key, err := ioutil.ReadFile(localKeyFile)
	if err != nil {
		return nil, err
	}
	return key, nil
}
