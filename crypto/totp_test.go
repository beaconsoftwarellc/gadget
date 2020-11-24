package crypto

import (
    "crypto/sha1"
    "encoding/base32"
    assert1 "github.com/stretchr/testify/assert"
    "os"
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
    type args struct {
        key    string
        step   int
        vary   int
        length int
    }
    tests := []struct {
        name    string
        args    args
        want    string
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := TOTP(tt.args.key, tt.args.step, tt.args.vary, tt.args.length)
            if (err != nil) != tt.wantErr {
                t.Errorf("TOTP() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("TOTP() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestTOTPCompare(t *testing.T) {
    type args struct {
        key       string
        step      time.Duration
        adjust    int
        length    int
        challenge string
    }
    tests := []struct {
        name    string
        args    args
        want    bool
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := TOTPCompare(tt.args.key, tt.args.step, tt.args.adjust, tt.args.length, tt.args.challenge)
            if (err != nil) != tt.wantErr {
                t.Errorf("TOTPCompare() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("TOTPCompare() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestTOTPCompareWithVariance(t *testing.T) {
    type args struct {
        key       string
        step      time.Duration
        length    int
        variance  uint
        challenge string
    }
    tests := []struct {
        name    string
        args    args
        want    bool
        wantErr bool
    }{
        // TODO: Add test cases.
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := TOTPCompareWithVariance(tt.args.key, tt.args.step, tt.args.length, tt.args.variance, tt.args.challenge)
            if (err != nil) != tt.wantErr {
                t.Errorf("TOTPCompareWithVariance() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("TOTPCompareWithVariance() got = %v, want %v", got, tt.want)
            }
        })
    }
}

func TestQRCode(t *testing.T) {
    assert := assert1.New(t)
    f, _ := os.OpenFile("qrcode.png", os.O_CREATE|os.O_TRUNC|os.O_WRONLY, os.ModePerm)
    key, _ := NewOTPKey()
    bytes, err := GenerateTOTPQRCodePNG(key, "eCatholic", "anthony@ecatholic.com", 30*time.Second, 6)
    assert.NoError(err)
    _, err = f.Write(bytes)
    assert.NoError(err)
    f.Close()
}