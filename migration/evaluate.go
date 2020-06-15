package migration

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"fmt"
	"hash"
	"net/url"
	"time"

	"google.golang.org/protobuf/proto"
)

var (
	hashes = map[Payload_Algorithm]func() hash.Hash{
		Payload_ALGORITHM_UNSPECIFIED: sha1.New, // default
		Payload_ALGORITHM_SHA1:        sha1.New,
		Payload_ALGORITHM_SHA256:      sha256.New,
		Payload_ALGORITHM_SHA512:      sha512.New,
		Payload_ALGORITHM_MD5:         md5.New,
	}
	digits = map[Payload_DigitCount]int{
		Payload_DIGIT_COUNT_UNSPECIFIED: 1e6, // default
		Payload_DIGIT_COUNT_SIX:         1e6,
		Payload_DIGIT_COUNT_EIGHT:       1e8,
	}
)

// now function for testing purposes
var now = time.Now

func (op *Payload_OtpParameters) evaluate(c int64) int {
	h := hmac.New(hashes[op.Algorithm], op.Secret)
	binary.Write(h, binary.BigEndian, c)
	hash := h.Sum(nil)
	off := hash[h.Size()-1] & 15
	header := binary.BigEndian.Uint32(hash[off:]) & (1<<31 - 1)
	return int(header) % digits[op.Digits]
}

// Evaluate OTP parameters
func (op *Payload_OtpParameters) Evaluate() int {
	switch op.Type {
	case Payload_OTP_TYPE_HOTP:
		return op.evaluate(op.Counter) // TODO increment counter
	case Payload_OTP_TYPE_TOTP:
		return op.evaluate(now().Unix() / 30) // default period 30s
	}
	return 0
}

// Evaluate otpauth-migration URL
func Evaluate(u *url.URL) error {
	data, err := dataQuery(u)
	if err != nil {
		return err
	}
	var p Payload
	if err := proto.Unmarshal(data, &p); err != nil {
		return err
	}
	for _, op := range p.OtpParameters {
		fmt.Printf("%s %06d", op.Name, op.Evaluate())
	}
	return nil
}