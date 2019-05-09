package certificate

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"

	"math/rand"
	"os"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func randomString(n int) string {
	buffer := make([]byte, n)
	for i := range buffer {
		buffer[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(buffer)
}

func log(infoType string, info string, log string) {
	currentTime := "[" + time.Now().Format("2006-01-02 15:04:05") + "] "
	fmt.Println(currentTime + "[" + infoType + "] " + info)
	file, _error := os.OpenFile(log, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if _error != nil {
		fmt.Println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _, _error := file.Write([]byte(currentTime + "[" + infoType + "] " + info)); _error != nil {
		fmt.Println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	if _error := file.Close(); _error != nil {
		fmt.Println(currentTime + "[ERROR] " + _error.Error() + "\n")
		return
	}
	return
}

//GenerateCertificate takes in a logpath and certificatPath and generates a public&private key
func GenerateCertificate(certificatePath, logPath string) {
	certificateConfig := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Organization:  []string{randomString(rand.Int())},
			Country:       []string{randomString(rand.Int())},
			Province:      []string{randomString(rand.Int())},
			Locality:      []string{randomString(rand.Int())},
			StreetAddress: []string{randomString(rand.Int())},
			PostalCode:    []string{randomString(rand.Int())},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	var Reader io.Reader
	privateKey, err := rsa.GenerateKey(Reader, 2048)
	if err != nil {
		log("create privateKey failed", err.Error(), logPath)
		return
	}
	publicKey := &privateKey.PublicKey
	certificate, err := x509.CreateCertificate(Reader, certificateConfig, certificateConfig, publicKey, privateKey)
	if err != nil {
		log("create certificate failed", err.Error(), logPath)
		return
	}

	// Public key
	certificatePublicOut, err := os.Create(certificatePath + "certificate.crt")
	if err != nil {
		log("create certificate failed", err.Error(), logPath)
		return
	}
	err = pem.Encode(certificatePublicOut, &pem.Block{Type: "CERTIFICATE", Bytes: certificate})
	if err != nil {
		log("create certificate failed", err.Error(), logPath)
		return
	}
	certificatePublicOut.Close()
	println("written certificate.pem\n")

	// Private key
	certificatePrivateOut, err := os.OpenFile(certificatePath+"certificate.key", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log("create certificate failed", err.Error(), logPath)
		return
	}
	err = pem.Encode(certificatePrivateOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})
	if err != nil {
		log("create certificate failed", err.Error(), logPath)
		return
	}
	certificatePrivateOut.Close()
	println("written certificatePrivateOut.pem\n")
}
