package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/adwski/vidi/internal/api/user/auth"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPICmd_CreateServiceToken(t *testing.T) {
	buf := bytes.Buffer{}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"api", "create-service-token", "-n", "testsvc", "-s", "jwtsecret", "-e", "1h"})
	err := rootCmd.Execute()

	require.NoError(t, err)

	tokenStr := buf.String()
	require.NotEmpty(t, tokenStr)

	token, err := jwt.ParseWithClaims(tokenStr, &auth.Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("jwtsecret"), nil
	})
	require.NoError(t, err)
	require.NotNil(t, token)

	assert.Equal(t, token.Claims.(*auth.Claims).Role, "service")
	assert.Equal(t, token.Claims.(*auth.Claims).Name, "testsvc")
}

//nolint:lll // mp4 sample file dump
const testFileDump = `ftyp: [isom iso2 avc1 mp41]
segmented: false
timescale: 15360 units per second
duration: 10s

segmentation info:
segment points (10) with 1s duration (err: <nil>): [{1 0 0} {31 15360 15360} {61 30720 30720} {91 46080 46080} {121 61440 61440} {151 76800 76800} {181 92160 92160} {211 107520 107520} {241 122880 122880} {271 138240 138240}]
TrackID: 1, type: vide, sampleCount: [300]
Codec info: &{avc1.64001f 0} (err: <nil>)
Segment intervals (err: <nil>): [{1 30} {31 60} {61 90} {91 120} {121 150} {151 180} {181 210} {211 240} {241 270} {271 299}]
TrackID: 2, type: soun, sampleCount: [472]
Codec info: &{mp4a.40.2 48000} (err: <nil>)
Segment intervals (err: <nil>): [{1 47} {48 94} {95 141} {142 188} {189 235} {236 282} {283 329} {330 375} {376 422} {423 471}]
TrackID: 3, type: tmcd, sampleCount: [1]
Codec info: <nil> (err: could not find proper av box in stsd)
Segment intervals (err: <nil>): [{1 1} {2 1} {2 1} {2 1} {2 1} {2 1} {2 1} {2 1} {2 1} {2 0}]

Codecs are supported!
`

func TestMP4Cmd_Dump(t *testing.T) {
	tmp := os.TempDir()
	testFileName := tmp + "/test.mp4"

	err := copyFile(testFileName, "../../testfiles/test_seq_h264_high.mp4")
	require.NoError(t, err)

	buf := bytes.Buffer{}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"mp4", "dump", "-f", testFileName, "-s", "1s"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	require.Equal(t, testFileDump, out)
}

func TestMP4Cmd_Segment(t *testing.T) {
	tmp := t.TempDir()
	testFileName := tmp + "/test.mp4"

	err := copyFile(testFileName, "../../testfiles/test_seq_h264_high.mp4")
	require.NoError(t, err)

	buf := bytes.Buffer{}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"mp4", "segment", "-f", testFileName, "-o", tmp, "-s", "1s"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "processing is done")

	dir, err := os.ReadDir(tmp)
	require.NoError(t, err)

	var (
		inits, other, soun, vide, manifest int
	)
	for _, d := range dir {
		switch {
		case strings.Contains(d.Name(), "_init.mp4"):
			inits++
		case strings.HasPrefix(d.Name(), "vide") && strings.HasSuffix(d.Name(), ".m4s"):
			vide++
		case strings.HasPrefix(d.Name(), "soun") && strings.HasSuffix(d.Name(), ".m4s"):
			soun++
		case d.Name() == "manifest.mpd":
			manifest++
		default:
			other++
		}
	}
	require.Equal(t, 2, inits, "two init segments")
	require.Equal(t, 10, vide, "10 video segments")
	require.Equal(t, 10, soun, "10 audio segments")
	require.Equal(t, 1, other, "1 source file")
	require.Equal(t, 1, manifest, "1 mpd file")
}

func TestMP4Cmd_SegmentNoFile(t *testing.T) {
	tmp := t.TempDir()
	testFileName := tmp + "/test.mp4"

	buf := bytes.Buffer{}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"mp4", "segment", "-f", testFileName, "-o", tmp, "-s", "1s"})
	err := rootCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "cannot open file")
}

func TestMP4Cmd_SegmentInvalidFile(t *testing.T) {
	tmp := t.TempDir()
	testFileName := tmp + "/test.mp4"

	err := os.WriteFile(testFileName, []byte("qwqwdqsdsad"), 0600)
	require.NoError(t, err)

	buf := bytes.Buffer{}
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)
	rootCmd.SetArgs([]string{"mp4", "segment", "-f", testFileName, "-o", tmp, "-s", "1s"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	out := buf.String()
	require.Contains(t, out, "error processing file")
}

func copyFile(dst string, src string) error {
	fSrc, err := os.Open(src)
	if err != nil {
		return err //nolint:wrapcheck // unnecessary
	}
	fDst, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err //nolint:wrapcheck // unnecessary
	}
	_, err = fSrc.WriteTo(fDst)
	return err //nolint:wrapcheck // unnecessary
}
