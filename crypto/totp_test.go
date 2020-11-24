package crypto

import (
    "crypto/sha1"
    "encoding/base32"
    assert1 "github.com/stretchr/testify/assert"
    "testing"
    "time"
)

func TestDynamicTruncate(t *testing.T) {
    type args struct {
        hmacResult []byte
        digits     int
    }
    tests := []struct {
        name string
        args args
        want string
    }{
        {
            "test from RFC",
            args{[]byte{0x1f, 0x86, 0x98, 0x69, 0x0e, 0x02, 0xca, 0x16, 0x61, 0x85, 0x50, 0xef, 0x7f, 0x19, 0xda, 0x8e, 0x94, 0x5b, 0x55, 0x5a}, 6},
            "872921",
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := DynamicTruncate(tt.args.hmacResult, tt.args.digits); got != tt.want {
                t.Errorf("DynamicTruncate() = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestHOTP(t *testing.T) {
    type args struct {
        key     string
        counter uint64
        length  int
    }
    tests := []struct {
        name    string
        args    args
        want    string
        wantErr bool
    }{
        {
            "works as expected 6",
            args{
                "AAAQEAYEAUDAOCAJBIFQYDIOB4IBCEQT",
                0,
                6,
            },
            "858575",
            false,
        },
        {
            "works as expected 7",
            args{
                "AAAQEAYEAUDAOCAJBIFQYDIOB4IBCEQT",
                0,
                8,
            },
            "67858575",
            false,
        },
        {
            "works as expected 8",
            args{
                "AAAQEAYEAUDAOCAJBIFQYDIOB4IBCEQT",
                0,
                8,
            },
            "67858575",
            false,
        },
        {
            "out of bounds fails low",
            args{
                "AAAQEAYEAUDAOCAJBIFQYDIOB4IBCEQT",
                0,
                5,
            },
            "",
            true,
        },
        {
            "out of bounds high fails",
            args{
                "AAAQEAYEAUDAOCAJBIFQYDIOB4IBCEQT",
                0,
                9,
            },
            "",
            true,
        },
        {
            "works as expected counter advance",
            args{
                "AAAQEAYEAUDAOCAJBIFQYDIOB4IBCEQT",
                4,
                6,
            },
            "455505",
            false,
        },
        {
            "bad data fails",
            args{
                "AAAaSQEAYEAUDA",
                0,
                6,
            },
            "",
            true,
        },
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := HOTP(tt.args.key, tt.args.counter, tt.args.length)
            if (err != nil) != tt.wantErr {
                t.Errorf("HOTP() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("HOTP() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestNewOTPKey(t *testing.T) {
    assert := assert1.New(t)
    encodedKey, err := NewOTPKey()
    assert.NoError(err)
    assert.Len(encodedKey, 32)
    key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(encodedKey)
    assert.NoError(err)
    assert.Len(key, sha1.Size)
}

func TestTOTP(t *testing.T) {
    assert := assert1.New(t)
    key, err := NewOTPKey()
    assert.NoError(err)
    _, err = TOTP(key, 30*time.Second, 0, 5)
    assert.Error(err)
    _, err = TOTP(key, 30*time.Second, 0, 9)
    assert.Error(err)
    totp, err := TOTP(key, 30*time.Second, 0, 6)
    assert.NoError(err)
    assert.Len(totp, 6)
    totp1, err := TOTP(key, 30*time.Second, 2, 6)
    assert.NoError(err)
    assert.Len(totp1, 6)
    assert.NotEqual(totp, totp1)
}

func TestTOTPCompare(t *testing.T) {
    assert := assert1.New(t)
    key, _ := NewOTPKey()
    totp, _ := TOTP(key, 30*time.Second, 0, 6)
    eq, err := TOTPCompare(key, 30*time.Second, 0, 6, totp)
    assert.NoError(err)
    assert.True(eq)

    totp, _ = TOTP(key, 30*time.Second, 1, 6)
    eq, err = TOTPCompare(key, 30*time.Second, 1, 6, totp)
    assert.NoError(err)
    assert.True(eq)

    totp, _ = TOTP(key, 30*time.Second, 1, 6)
    eq, err = TOTPCompare(key, 30*time.Second, -1, 6, totp)
    assert.NoError(err)
    assert.False(eq)
}

func TestTOTPCompareWithVariance(t *testing.T) {
    assert := assert1.New(t)
    key, _ := NewOTPKey()
    totp, _ := TOTP(key, 30*time.Second, 0, 6)
    eq, err := TOTPCompareWithVariance(key, 30*time.Second, 6, 0, totp)
    assert.NoError(err)
    assert.True(eq)

    totp, _ = TOTP(key, 30*time.Second, 2, 6)
    eq, err = TOTPCompareWithVariance(key, 30*time.Second, 6, 1, totp)
    assert.NoError(err)
    assert.False(eq)

    eq, err = TOTPCompareWithVariance(key, 30*time.Second, 6, 2, totp)
    assert.NoError(err)
    assert.True(eq)

    totp, _ = TOTP(key, 30*time.Second, -1, 6)
    eq, err = TOTPCompareWithVariance(key, 30*time.Second, 6, 2, totp)
    assert.NoError(err)
    assert.True(eq)
}

func TestGenerateTOTPURI(t *testing.T) {
    assert := assert1.New(t)
    key := "RTR62KM24TFDNICOUL7DBTLMJS42E3UZ"
    issuer := "Test Issuer"
    user := "user@test.com"
    period := 30*time.Second
    length := 6
    actual := GenerateTOTPURI(key, issuer, user, period, length)
    expected := "otpauth://totp/user%40test.com?secret=RTR62KM24TFDNICOUL7DBTLMJS42E3UZ&issuer=Test+Issuer&algorithm=SHA1&digits=6&period=30"
    assert.Equal(actual, expected)
}

func TestHOTPCompare(t *testing.T) {
    assert := assert1.New(t)
    key, _ := NewOTPKey()
    hotp, _ := HOTP(key, 0, 6)
    actual, err := HOTPCompare(key, 0, 6, hotp)
    assert.NoError(err)
    assert.True(actual)

    actual, err = HOTPCompare(key, 0, 5, hotp)
    assert.Error(err)
    assert.False(actual)

    hotp, _ = HOTP(key, 4, 6)
    actual, err = HOTPCompare(key, 4, 6, hotp)
    assert.NoError(err)
    assert.True(actual)

    hotp, _ = HOTP(key, 2, 6)
    actual, err = HOTPCompare(key, 4, 6, hotp)
    assert.NoError(err)
    assert.False(actual)
}