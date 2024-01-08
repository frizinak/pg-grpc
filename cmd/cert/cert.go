package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func ex(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func atomicWrite(path string, cb func(io.Writer) error) error {
	stamp := strconv.FormatInt(time.Now().UnixNano(), 36)
	rnd := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, rnd)
	if err != nil {
		panic(err)
	}

	tmp := fmt.Sprintf(
		"%s.%s-%s.tmp",
		path,
		stamp,
		base64.RawURLEncoding.EncodeToString(rnd),
	)

	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	err = cb(f)
	f.Close()
	if err != nil {
		os.Remove(tmp)
		return err
	}

	return os.Rename(tmp, path)
}

func priv(path string) (*rsa.PrivateKey, error) {
	var pk *rsa.PrivateKey

	f, err := os.Open(path)
	if err == nil {
		d, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(d)
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	return pk, atomicWrite(path, func(w io.Writer) error {
		rsapk, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return err
		}

		pk = rsapk

		pkb := x509.MarshalPKCS1PrivateKey(rsapk)

		return pem.Encode(w, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: pkb})
	})
}

func cert(path string, cert, ca *x509.Certificate, certKey crypto.PublicKey, caKey crypto.PrivateKey) error {
	_, err := os.Stat(path)
	if err == nil {
		return nil
	}

	if !os.IsNotExist(err) {
		return err
	}

	cab, err := x509.CreateCertificate(rand.Reader, cert, ca, certKey, caKey)
	ex(err)

	return atomicWrite(path, func(w io.Writer) error {
		return pem.Encode(w, &pem.Block{Type: "CERTIFICATE", Bytes: cab})
	})
}

func main() {
	domain := os.Args[1]
	dir := os.Args[2]

	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(5468645684645123),
		Subject:               pkix.Name{Organization: []string{"CA PGGRPC Inc."}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	c := &x509.Certificate{
		SerialNumber: big.NewInt(1234),
		Subject:      pkix.Name{Organization: []string{"PGGRPC Inc."}},
		DNSNames:     []string{domain},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(5, 0, 0),
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	capk, err := priv(filepath.Join(dir, "ca.key"))
	ex(err)
	cpk, err := priv(filepath.Join(dir, "cert.key"))
	ex(err)

	ex(cert(filepath.Join(dir, "ca.pem"), ca, ca, &capk.PublicKey, capk))
	ex(cert(filepath.Join(dir, "cert.pem"), c, ca, &cpk.PublicKey, capk))
}
