package mp4

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDump(t *testing.T) {
	buf := bytes.NewBuffer(make([]byte, 0, 10000))

	Dump(buf, "../../testfiles/test_seq_h264_high.mp4", time.Second)

	require.NotZero(t, buf.Len())
	str := buf.String()

	assert.Contains(t, str, "timescale: 15360 units per second")
	assert.Contains(t, str, "duration: 10s")
	assert.Contains(t, str, "TrackID: 1, type: vide, sampleCount: [300]")
	assert.Contains(t, str, "TrackID: 2, type: soun, sampleCount: [472]")
	assert.Contains(t, str, "Codecs are supported!")
}
