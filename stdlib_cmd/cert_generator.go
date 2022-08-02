package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	utls "github.com/getlantern/utls"
	"net"
	"time"
)

type certificateGenerator struct {
	ca             tls.Certificate
	caX509         *x509.Certificate
	fallbackTarget string
}

func newCertGenerator(ca tls.Certificate, fallbackTarget string) (*certificateGenerator, error) {
	caX509, err := x509.ParseCertificate(ca.Certificate[0])
	if err != nil {
		return nil, err
	}
	return &certificateGenerator{ca: ca, caX509: caX509, fallbackTarget: fallbackTarget}, nil
}

func (cg *certificateGenerator) genChildCert(ips, names []string) (*tls.Certificate, error) {
	private, cab, err := cg.genCertBytes(ips, names)
	if err != nil {
		return nil, err
	}

	return &tls.Certificate{
		Certificate: [][]byte{cab},
		PrivateKey:  private,
	}, nil
}

func (cg *certificateGenerator) genChildCertUTLS(ips, names []string) (*utls.Certificate, error) {
	private, cab, err := cg.genCertBytes(ips, names)
	if err != nil {
		return nil, err
	}

	return &utls.Certificate{
		Certificate: [][]byte{cab},
		PrivateKey:  private,
	}, nil
}

func (cg *certificateGenerator) genCertBytes(ips []string, names []string) (*rsa.PrivateKey, []byte, error) {
	s, _ := rand.Prime(rand.Reader, 128)

	template := &x509.Certificate{
		SerialNumber:          s,
		Subject:               pkix.Name{Organization: []string{"mitmproxy"}},
		Issuer:                pkix.Name{Organization: []string{"mitmproxy"}},
		NotBefore:             time.Now().AddDate(-1, 0, 0),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		BasicConstraintsValid: true,
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
	}
	if ips != nil {
		is := make([]net.IP, 0)
		for _, i := range ips {
			is = append(is, net.ParseIP(i))
		}
		template.IPAddresses = is
	}
	if names != nil {
		template.DNSNames = names
	}

	private := cg.ca.PrivateKey.(*rsa.PrivateKey)

	certP, _ := x509.ParseCertificate(cg.ca.Certificate[0])
	public := certP.PublicKey.(*rsa.PublicKey)

	cab, err := x509.CreateCertificate(rand.Reader, template, cg.caX509, public, private)
	if err != nil {
		return nil, nil, err
	}
	return private, cab, nil
}
