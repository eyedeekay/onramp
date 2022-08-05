package onramp

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cretz/bine/torutil"
)

// TLSKeys returns the TLS certificate and key for the given Garlic.
// if no TLS keys exist, they will be generated. They will be valid for
// the .b32.i2p domain.
func (g *Garlic) TLSKeys() (tls.Certificate, error) {
	keys, err := g.Keys()
	if err != nil {
		return tls.Certificate{}, err
	}
	base32 := keys.Addr().Base32()
	return TLSKeys(base32)
}

// TLSKeys returns the TLS certificate and key for the given Onion.
// if no TLS keys exist, they will be generated. They will be valid for
// the .onion domain.
func (o *Onion) TLSKeys() (tls.Certificate, error) {
	keys, err := o.Keys()
	if err != nil {
		return tls.Certificate{}, err
	}
	onionService := torutil.OnionServiceIDFromPrivateKey(keys.PrivateKey)
	return TLSKeys(onionService)
}

// TLSKeys returns the TLS certificate and key for the given hostname.
func TLSKeys(tlsHost string) (tls.Certificate, error) {
	tlsCert := tlsHost + ".crt"
	tlsKey := tlsHost + ".pem"
	if err := CreateTLSCertificate(tlsHost); nil != err {
		return tls.Certificate{}, err
	}
	tlsKeystorePath, err := TLSKeystorePath()
	if err != nil {
		return tls.Certificate{}, err
	}
	tlsCertPath := filepath.Join(tlsKeystorePath, tlsCert)
	tlsKeyPath := filepath.Join(tlsKeystorePath, tlsKey)
	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		return cert, err
	}
	return cert, nil
}

// CreateTLSCertificate generates a TLS certificate for the given hostname,
// and stores it in the TLS keystore for the application. If the keys already
// exist, generation is skipped.
func CreateTLSCertificate(tlsHost string) error {
	tlsCertName := tlsHost + ".crt"
	tlsKeyName := tlsHost + ".pem"
	tlsKeystorePath, err := TLSKeystorePath()
	if err != nil {
		return err
	}
	tlsCert := filepath.Join(tlsKeystorePath, tlsCertName)
	tlsKey := filepath.Join(tlsKeystorePath, tlsKeyName)
	_, certErr := os.Stat(tlsCert)
	_, keyErr := os.Stat(tlsKey)
	if certErr != nil || keyErr != nil {
		if certErr != nil {
			fmt.Printf("Unable to read TLS certificate '%s'\n", tlsCert)
		}
		if keyErr != nil {
			fmt.Printf("Unable to read TLS key '%s'\n", tlsKey)
		}

		if err := createTLSCertificate(tlsHost); nil != err {
			return err
		}
	}

	return nil
}

func createTLSCertificate(host string) error {
	fmt.Println("Generating TLS keys. This may take a minute...")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}

	tlsCert, err := NewTLSCertificate(host, priv)
	if nil != err {
		return err
	}
	privStore, err := TLSKeystorePath()
	if nil != err {
		return err
	}

	certFile := filepath.Join(privStore, host+".crt")
	// save the TLS certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %s", host+".crt", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: tlsCert})
	certOut.Close()
	fmt.Printf("\tTLS certificate saved to: %s\n", host+".crt")

	// save the TLS private key
	privFile := filepath.Join(privStore, host+".pem")
	keyOut, err := os.OpenFile(privFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %v", privFile, err)
	}
	secp384r1, err := asn1.Marshal(asn1.ObjectIdentifier{1, 3, 132, 0, 34}) // http://www.ietf.org/rfc/rfc5480.txt
	pem.Encode(keyOut, &pem.Block{Type: "EC PARAMETERS", Bytes: secp384r1})
	ecder, err := x509.MarshalECPrivateKey(priv)
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecder})
	pem.Encode(keyOut, &pem.Block{Type: "CERTIFICATE", Bytes: tlsCert})

	keyOut.Close()
	fmt.Printf("\tTLS private key saved to: %s\n", privFile)

	// CRL
	crlFile := filepath.Join(privStore, host+".crl")
	crlOut, err := os.OpenFile(crlFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open %s for writing: %s", crlFile, err)
	}
	crlcert, err := x509.ParseCertificate(tlsCert)
	if err != nil {
		return fmt.Errorf("Certificate with unknown critical extension was not parsed: %s", err)
	}

	now := time.Now()
	revokedCerts := []pkix.RevokedCertificate{
		{
			SerialNumber:   crlcert.SerialNumber,
			RevocationTime: now,
		},
	}

	crlBytes, err := crlcert.CreateCRL(rand.Reader, priv, revokedCerts, now, now)
	if err != nil {
		return fmt.Errorf("error creating CRL: %s", err)
	}
	_, err = x509.ParseDERCRL(crlBytes)
	if err != nil {
		return fmt.Errorf("error reparsing CRL: %s", err)
	}
	pem.Encode(crlOut, &pem.Block{Type: "X509 CRL", Bytes: crlBytes})
	crlOut.Close()
	fmt.Printf("\tTLS CRL saved to: %s\n", crlFile)

	return nil
}

// NewTLSCertificate generates a new TLS certificate for the given hostname,
// returning it as bytes.
func NewTLSCertificate(host string, priv *ecdsa.PrivateKey) ([]byte, error) {
	return NewTLSCertificateAltNames(priv, host)
}

// NewTLSCertificateAltNames generates a new TLS certificate for the given hostname,
// and a list of alternate names, returning it as bytes.
func NewTLSCertificateAltNames(priv *ecdsa.PrivateKey, hosts ...string) ([]byte, error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(5 * 365 * 24 * time.Hour)
	host := ""
	if len(hosts) > 0 {
		host = hosts[0]
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization:       []string{"I2P Anonymous Network"},
			OrganizationalUnit: []string{"I2P"},
			Locality:           []string{"XX"},
			StreetAddress:      []string{"XX"},
			Country:            []string{"XX"},
			CommonName:         host,
		},
		NotBefore:          notBefore,
		NotAfter:           notAfter,
		SignatureAlgorithm: x509.ECDSAWithSHA512,

		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		DNSNames:              hosts[1:],
	}

	hosts = append(hosts, strings.Split(host, ",")...)
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return nil, err
	}

	return derBytes, nil
}
