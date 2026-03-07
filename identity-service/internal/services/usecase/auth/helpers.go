package auth

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	toddlerr "github.com/beka-birhanu/toddler/error"
	"github.com/beka-birhanu/toddler/status"
)

func normalizePhone(phone string) string {
	phone = strings.TrimSpace(phone)
	if strings.HasPrefix(phone, "0") {
		phone = "+251" + phone[1:]
	}
	return phone
}

func GenerateOTP(length int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(length)), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", length, n), nil
}

func invalidCredentialsError() error {
	return &toddlerr.Error{
		PublicStatusCode:  status.Unauthorized,
		ServiceStatusCode: status.Unauthorized,
		PublicMessage:     "Invalid username or password",
		ServiceMessage:    "invalid credentials",
	}
}

func lockedError() error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     "Too many attempts. Please try again later",
		ServiceMessage:    "otp locked",
	}
}

func otpNotVerifiedError() error {
	return &toddlerr.Error{
		PublicStatusCode:  status.BadRequest,
		ServiceStatusCode: status.BadRequest,
		PublicMessage:     "OTP not verified",
		ServiceMessage:    "otp has not been verified for this random",
	}
}

func mobileOtpKey(mobile string) string {
	return "otp:mobile:" + mobile
}
