package library

// Alternative key generation:
// Private key:
// openssl genrsa -out app.key 2048
// Public key:
// openssl rsa -in app.key -pubout > app.key.pub

import (
  "crypto/rand"
  "crypto/rsa"
  "crypto/x509"
  "encoding/pem"
  "errors"
  "io/ioutil"
  "log"
  "net/http"
  "os"
  "time"
  "github.com/SermoDigital/jose/jws"
  "github.com/SermoDigital/jose/crypto"
)

func InitJwtKeys() error {
   _, errPriv := os.Stat("app.rsa");
   _, errPub := os.Stat("app.rsa.pub");
  if os.IsNotExist(errPriv) || os.IsNotExist(errPub) {
    log.Println("Create JWT RSA keys...")

    // generate 2048 bit rsa key
    reader := rand.Reader
    key, err := rsa.GenerateKey(reader, 2048)
    if err != nil {
      return err
    }

    // save private key
    privFile, err := os.OpenFile("app.rsa", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
    if err != nil {
      return err
    }
    defer privFile.Close()

    var privateKey = &pem.Block {
      Type:  "PRIVATE RSA KEY",
      Bytes: x509.MarshalPKCS1PrivateKey(key),
    }
    err = pem.Encode(privFile, privateKey)
    if err != nil {
      return err
    }

    // save public key
    asn1Bytes, err := x509.MarshalPKIXPublicKey(&(key.PublicKey))
    if err != nil {
      return err
    }

    var pubkey = &pem.Block{
      Type:  "PUBLIC RSA KEY",
      Bytes: asn1Bytes,
    }

    pubFile, err := os.OpenFile("app.rsa.pub", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
    if err != nil {
      return err
    }
    defer pubFile.Close()

    err = pem.Encode(pubFile, pubkey)
    if err != nil {
      return err
    }

  } else {
    log.Println("Use existing JWT RSA keys...")
  }
  return nil
}

func CreateJwtToken(subject, name string) (string, error) {

  expires := time.Now().Add(time.Duration(24) * time.Hour)

  claims := jws.Claims{}
  claims.SetExpiration(expires)
  claims.SetIssuedAt(time.Now())
  claims.SetSubject(subject)
  claims.Set("name", name)

  bytes, err := ioutil.ReadFile("app.rsa")
  if err != nil {
    return "", err
  }

  rsaPrivate, err := crypto.ParseRSAPrivateKeyFromPEM(bytes)
  if err != nil {
    return "", err
  }
  jwt := jws.NewJWT(claims, crypto.SigningMethodRS256)

  b, err := jwt.Serialize(rsaPrivate)
  if err != nil {
    return "", err
  }
  return string(b), nil
}

func ValidateJwtAndGetSubject(r *http.Request) (string, error) {

  bytes, err := ioutil.ReadFile("app.rsa.pub")
  if err != nil {
    return "", err
  }
  rsaPublic, err := crypto.ParseRSAPublicKeyFromPEM(bytes)
  if err != nil {
    return "", err
  }

  jwt, err := jws.ParseJWTFromRequest(r)
  if err != nil {
    return "", err
  }

  // Validate token
  if err = jwt.Validate(rsaPublic, crypto.SigningMethodRS256); err != nil {
    return "", err
  }

  if t, set := jwt.Claims().Expiration(); ! set || t.Before(time.Now ()) {
    return "", err
  }

  subject, ok := jwt.Claims().Subject()
  if ! ok {
    return "", errors.New("JWT has no subject")
  }
  return subject, nil
}
