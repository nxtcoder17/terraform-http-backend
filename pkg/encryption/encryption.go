package encryption

type Cipher interface {
	Encrypt(content []byte) ([]byte, error)
	Decrypt(encrypted []byte) ([]byte, error)
}
