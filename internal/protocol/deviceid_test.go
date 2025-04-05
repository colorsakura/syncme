package protocol

import "testing"

func TestDeviceID(t *testing.T) {
	id, err := NewDeviceID([]byte(""))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(id)
}
