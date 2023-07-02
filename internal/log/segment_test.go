package log

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	api "github.com/yapatta/proglog/api/v1"
	"google.golang.org/protobuf/proto"
)

func TestSegment(t *testing.T) {
	dir, err := os.MkdirTemp("", "segment-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	want := &api.Record{Value: []byte("hello world")}

	c := Config{}
	c.Segment.MaxStoreBytes = 1024
	c.Segment.MaxIndexBytes = entWidth * 3

	s, err := newSegment(dir, 16, c)
	require.NoError(t, err)
	require.Equal(t, uint64(16), s.nextOffset)
	require.False(t, s.IsMaxed())

	for i := uint64(0); i < 3; i++ {
		off, err := s.Append(want)
		require.NoError(t, err)
		require.Equal(t, 16+i, off)

		got, err := s.Read(off)
		require.NoError(t, err)
		require.Equal(t, want.Value, got.Value)
	}

	_, err = s.Append(want)
	require.Equal(t, io.EOF, err)

	// MEMO: no space in index
	require.True(t, s.IsMaxed())
	require.NoError(t, s.Close())

	p, _ := proto.Marshal(want)
	c.Segment.MaxStoreBytes = uint64(len(p)+lenWidth) * 4
	c.Segment.MaxIndexBytes = 1024
	// MEMO: renew the existing segment
	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	// MEMO: store has no space
	require.True(t, s.IsMaxed())

	require.NoError(t, s.Remove())

	s, err = newSegment(dir, 16, c)
	require.NoError(t, err)
	require.False(t, s.IsMaxed())
	require.NoError(t, s.Close())
}
