package global

import "fmt"
import "os"
import "github.com/m3ng9i/go-utils/tls"


func makeCertFiles(cert, key string, overwrite bool) error {
    if cert == "" || key == "" {
        return fmt.Errorf("Both certificate path and key path should be provided")
    }

    // if the certificate or private key is exist, return an error
    if !overwrite {
        certExist := false
        keyExist := false
        if _, err := os.Stat(cert); err == nil {
            certExist = true
        }
        if _, err := os.Stat(key); err == nil {
            keyExist = true
        }
        if certExist && keyExist {
            return fmt.Errorf("Certificate and private key are all exist, remove them and try again.")
        }
        if certExist {
            return fmt.Errorf("Certificate is exist, remove it and try again.")
        }
        if keyExist {
            return fmt.Errorf("Private key is exist, remove it and try again.")
        }
    }

    // generate certificate and private key

    option := tls.DefaultCertOption()
    option.PublicKey    = cert
    option.PrivateKey   = key
    option.Organization = "RanServer"

    return tls.MakeCert(option)
}

