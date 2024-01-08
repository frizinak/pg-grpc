package main

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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

func priv(path string) (*ecdsa.PrivateKey, error) {
	var pk *ecdsa.PrivateKey

	f, err := os.Open(path)
	if err == nil {
		d, err := io.ReadAll(f)
		f.Close()
		if err != nil {
			return nil, err
		}
		block, _ := pem.Decode(d)
		return x509.ParseECPrivateKey(block.Bytes)
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	return pk, atomicWrite(path, func(w io.Writer) error {
		ecpk, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return err
		}

		pk = ecpk

		pkb, err := x509.MarshalECPrivateKey(ecpk)
		if err != nil {
			return err
		}

		return pem.Encode(w, &pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: pkb})
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
	dir := "./"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}

	c := &x509.Certificate{
		SerialNumber:          big.NewInt(5468645684645123),
		Subject:               pkix.Name{Organization: []string{"PGGRPC Inc."}},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(5, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	pk, err := priv(filepath.Join(dir, "cert.key"))
	ex(err)
	ex(cert(filepath.Join(dir, "cert.pem"), c, c, &pk.PublicKey, pk))
}
