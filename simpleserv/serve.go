package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"
)

type ProtectedHandler struct {
	H    http.Handler
	User string
	Pass string
}

func (h *ProtectedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.User != "" && h.Pass != "" {
		w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

		user, pass, authOK := r.BasicAuth()
		if authOK == false {
			http.Error(w, "Not authorized", 401)
			return
		}

		if user != h.User || pass != h.Pass {
			http.Error(w, "Not authorized", 401)
			return
		}
	}

	h.H.ServeHTTP(w, r)
}

// From https://golang.org/src/net/http/server.go
// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

// Adapted from https://gist.github.com/shivakar/cd52b5594d4912fbeb46
func GenX509KeyPair() (tls.Certificate, error) {
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: big.NewInt(now.Unix()),
		Subject: pkix.Name{
			Organization: []string{"My Org"},
		},
		NotBefore:             now,
		NotAfter:              now.AddDate(0, 1, 0), // Valid for one month
		SubjectKeyId:          []byte{113, 117, 105, 99, 107, 115, 101, 114, 118, 101},
		BasicConstraintsValid: true,
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		KeyUsage: x509.KeyUsageKeyEncipherment |
			x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf(
			"Failed to generate private key: %w", err)
	}

	cert, err := x509.CreateCertificate(rand.Reader, template, template,
		priv.Public(), priv)
	if err != nil {
		return tls.Certificate{}, err
	}

	var outCert tls.Certificate
	outCert.Certificate = append(outCert.Certificate, cert)
	outCert.PrivateKey = priv

	return outCert, nil
}

func main() {
	var (
		port     = flag.Int("port", 8081, "port to serve on")
		user     = flag.String("u", "", "user to protect with")
		pass     = flag.String("p", "", "password to protect with")
		useHTTPS = flag.Bool("s", false, "enables https with random cert")
	)
	flag.Parse()

	mux := http.NewServeMux()

	mux.Handle("/", &ProtectedHandler{http.FileServer(http.Dir(".")), *user, *pass})

	server := &http.Server{
		Handler: mux,
		Addr:    fmt.Sprintf(":%d", *port),
	}
	log.Printf("Serving files in the current directory on port %d", *port)
	var err error
	if *useHTTPS {
		var cert tls.Certificate
		cert, err = GenX509KeyPair()
		if err != nil {
			log.Fatal(err)
		}

		server.TLSConfig = &tls.Config{
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
			NextProtos:               []string{"http/1.1"},
			Certificates:             []tls.Certificate{cert},
		}

		var lis net.Listener
		lis, err = net.Listen("tcp", server.Addr)
		if err == nil {
			tlsListener := tls.NewListener(
				tcpKeepAliveListener{lis.(*net.TCPListener)},
				server.TLSConfig)

			err = server.Serve(tlsListener)
		}
	} else {
		err = server.ListenAndServe()
	}

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
