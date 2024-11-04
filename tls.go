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
	"github.com/sirupsen/logrus"
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
	log.WithField("name", g.getName()).Debug("Getting TLS keys for Garlic service")
	keys, err := g.Keys()
	if err != nil {
		log.WithError(err).Error("Failed to get I2P keys")
		return tls.Certificate{}, err
	}
	base32 := keys.Addr().Base32()
	log.WithField("base32", base32).Debug("Retrieving TLS certificate for base32 address")
	return TLSKeys(base32)
}

// TLSKeys returns the TLS certificate and key for the given Onion.
// if no TLS keys exist, they will be generated. They will be valid for
// the .onion domain.
func (o *Onion) TLSKeys() (tls.Certificate, error) {
	log.WithField("name", o.getName()).Debug("Getting TLS keys for Onion service")
	keys, err := o.Keys()
	if err != nil {
		return tls.Certificate{}, err
	}
	onionService := torutil.OnionServiceIDFromPrivateKey(keys)
	log.WithField("onion_service", onionService).Debug("Retrieving TLS certificate for onion service")
	return TLSKeys(onionService)
}

// TLSKeys returns the TLS certificate and key for the given hostname.
func TLSKeys(tlsHost string) (tls.Certificate, error) {
	log.WithField("host", tlsHost).Debug("Getting TLS certificate and key")
	tlsCert := tlsHost + ".crt"
	tlsKey := tlsHost + ".pem"
	if err := CreateTLSCertificate(tlsHost); nil != err {
		log.WithError(err).Error("Failed to create TLS certificate")
		return tls.Certificate{}, err
	}
	tlsKeystorePath, err := TLSKeystorePath()
	if err != nil {
		log.WithError(err).Error("Failed to get TLS keystore path")
		return tls.Certificate{}, err
	}
	tlsCertPath := filepath.Join(tlsKeystorePath, tlsCert)
	tlsKeyPath := filepath.Join(tlsKeystorePath, tlsKey)

	log.WithFields(logrus.Fields{
		"cert_path": tlsCertPath,
		"key_path":  tlsKeyPath,
	}).Debug("Loading TLS certificate pair")

	cert, err := tls.LoadX509KeyPair(tlsCertPath, tlsKeyPath)
	if err != nil {
		log.WithError(err).Error("Failed to load TLS certificate pair")
		return cert, err
	}

	log.Debug("Successfully loaded TLS certificate and key")
	return cert, nil
}

// CreateTLSCertificate generates a TLS certificate for the given hostname,
// and stores it in the TLS keystore for the application. If the keys already
// exist, generation is skipped.
func CreateTLSCertificate(tlsHost string) error {
	log.WithField("host", tlsHost).Debug("Creating TLS certificate")
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
		log.WithFields(logrus.Fields{
			"cert_exists": certErr == nil,
			"key_exists":  keyErr == nil,
			"cert_path":   tlsCert,
			"key_path":    tlsKey,
		}).Debug("Certificate or key missing, generating new ones")
		if certErr != nil {
			log.WithField("path", tlsCert).Debug("TLS certificate not found")
			fmt.Printf("Unable to read TLS certificate '%s'\n", tlsCert)
		}
		if keyErr != nil {
			log.WithField("path", tlsKey).Debug("TLS key not found")
			fmt.Printf("Unable to read TLS key '%s'\n", tlsKey)
		}

		if err := createTLSCertificate(tlsHost); nil != err {
			log.WithError(err).Error("Failed to create TLS certificate")
			return err
		}
	} else {
		log.Debug("TLS certificate and key already exist")
	}

	return nil
}

func createTLSCertificate(host string) error {
	log.WithField("host", host).Debug("Generating new TLS certificate")
	fmt.Println("Generating TLS keys. This may take a minute...")
	priv, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		log.WithError(err).Error("Failed to generate private key")
		return err
	}

	tlsCert, err := NewTLSCertificate(host, priv)
	if nil != err {
		log.WithError(err).Error("Failed to create new TLS certificate")
		return err
	}
	privStore, err := TLSKeystorePath()
	if nil != err {
		log.WithError(err).Error("Failed to get keystore path")
		return err
	}

	certFile := filepath.Join(privStore, host+".crt")
	log.WithField("path", certFile).Debug("Saving TLS certificate")
	// save the TLS certificate
	certOut, err := os.Create(certFile)
	if err != nil {
		log.WithError(err).WithField("path", certFile).Error("Failed to create certificate file")
		return fmt.Errorf("failed to open %s for writing: %s", host+".crt", err)
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: tlsCert})
	certOut.Close()
	log.WithField("path", certFile).Debug("TLS certificate saved successfully")
	fmt.Printf("\tTLS certificate saved to: %s\n", host+".crt")

	// save the TLS private key
	privFile := filepath.Join(privStore, host+".pem")
	log.WithField("path", privFile).Debug("Saving TLS private key")
	keyOut, err := os.OpenFile(privFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.WithError(err).WithField("path", privFile).Error("Failed to create private key file")
		return fmt.Errorf("failed to open %s for writing: %v", privFile, err)
	}
	secp384r1, err := asn1.Marshal(asn1.ObjectIdentifier{1, 3, 132, 0, 34}) // http://www.ietf.org/rfc/rfc5480.txt
	if err != nil {
		log.WithError(err).Error("Failed to marshal EC parameters")
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PARAMETERS", Bytes: secp384r1})
	ecder, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		log.WithError(err).Error("Failed to marshal private key")
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "EC PRIVATE KEY", Bytes: ecder})
	pem.Encode(keyOut, &pem.Block{Type: "CERTIFICATE", Bytes: tlsCert})

	keyOut.Close()
	log.WithField("path", privFile).Debug("TLS private key saved successfully")
	fmt.Printf("\tTLS private key saved to: %s\n", privFile)

	// CRL
	crlFile := filepath.Join(privStore, host+".crl")
	log.WithField("path", crlFile).Debug("Creating CRL")
	crlOut, err := os.OpenFile(crlFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.WithError(err).WithField("path", crlFile).Error("Failed to create CRL file")
		return fmt.Errorf("failed to open %s for writing: %s", crlFile, err)
	}
	crlcert, err := x509.ParseCertificate(tlsCert)
	if err != nil {
		log.WithError(err).Error("Failed to parse certificate for CRL creation")
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
		log.WithError(err).Error("Failed to create CRL")
		return fmt.Errorf("error creating CRL: %s", err)
	}
	_, err = x509.ParseDERCRL(crlBytes)
	if err != nil {
		log.WithError(err).Error("Failed to validate generated CRL")
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
