package migration

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/binary"
	"hash"
	"time"
)

var (
	hashFunc = map[Payload_Algorithm]func() hash.Hash{
		Payload_ALGORITHM_UNSPECIFIED: sha1.New, // default
		Payload_ALGORITHM_SHA1:        sha1.New,
		Payload_ALGORITHM_SHA256:      sha256.New,
		Payload_ALGORITHM_SHA512:      sha512.New,
		Payload_ALGORITHM_MD5:         md5.New,
	}
	digitCount = map[Payload_DigitCount]int{
		Payload_DIGIT_COUNT_UNSPECIFIED: 1e6, // default
		Payload_DIGIT_COUNT_SIX:         1e6,
		Payload_DIGIT_COUNT_EIGHT:       1e8,
	}
	countFunc = map[Payload_OtpType]func(*Payload_OtpParameters) int64{
		Payload_OTP_TYPE_UNSPECIFIED: totp, // default
		Payload_OTP_TYPE_HOTP:        hotp,
		Payload_OTP_TYPE_TOTP:        totp,
	}
)

// now function for testing purposes
var now = time.Now

func hotp(op *Payload_OtpParameters) int64 {
	op.Counter++ // pre-increment rfc4226 section 7.2.
	return op.Counter
}

func totp(op *Payload_OtpParameters) int64 {
	return now().Unix() / 30
}

// Evaluate OTP parameters
func (op *Payload_OtpParameters) Evaluate() int {
	h := hmac.New(hashFunc[op.Algorithm], op.Secret)
	binary.Write(h, binary.BigEndian, countFunc[op.Type](op))
	hashed := h.Sum(nil)
	offset := hashed[h.Size()-1] & 15
	result := binary.BigEndian.Uint32(hashed[offset:]) & (1<<31 - 1)
	return int(result) % digitCount[op.Digits]
}
